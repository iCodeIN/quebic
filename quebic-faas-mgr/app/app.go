//    Copyright 2018 Tharanga Nilupul Thennakoon
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package app

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"quebic-faas/common"
	"quebic-faas/config"
	_messenger "quebic-faas/messenger"
	"quebic-faas/quebic-faas-mgr/components"
	mgrconfig "quebic-faas/quebic-faas-mgr/config"
	dao "quebic-faas/quebic-faas-mgr/dao"
	"quebic-faas/quebic-faas-mgr/db"
	dep "quebic-faas/quebic-faas-mgr/deployment"
	"quebic-faas/quebic-faas-mgr/deployment/docker_deployment"
	"quebic-faas/quebic-faas-mgr/deployment/kube_deployment"
	"quebic-faas/quebic-faas-mgr/httphandler"
	"quebic-faas/quebic-faas-mgr/logger"
	"quebic-faas/types"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	yaml "gopkg.in/yaml.v2"

	bolt "github.com/coreos/bbolt"
)

//App app container
type App struct {
	config     mgrconfig.AppConfig
	db         *bolt.DB
	messenger  _messenger.Messenger
	router     *mux.Router
	loggerUtil logger.Logger
	deployment dep.Deployment
}

//Start init app
func (app *App) Start() {

	log.Printf("quebic-faas-manager starting")

	//get db
	db, err := db.GetDb()
	defer db.Close()
	if err != nil {
		log.Fatalf("init failed. unable to connect db, error : %v", err)
	}
	app.db = db

	//setup root admin user to loging manager
	app.setupAdminUser()

	//setup deployment
	app.setupDeployment()

	//setup manager components
	app.setupManagerComponents()

	//setup router
	app.router = mux.NewRouter()
	app.setUpHTTPHandlers()

}

//DoConfiguration app configuration
func (app *App) DoConfiguration() {
	//setup configuration
	app.config.SetDefault()
	app.setupConfiguration()
}

func (app *App) setupConfiguration() {

	configFilePath := mgrconfig.GetConfigFilePath()

	readConfigJSON, err := ioutil.ReadFile(configFilePath)
	if err != nil {

		app.SaveConfiguration()

	} else {

		//if found config file set those configuration into App.config object
		savingConfig := mgrconfig.SavingConfig{}
		yaml.Unmarshal(readConfigJSON, &savingConfig)

		app.config.Auth = savingConfig.Auth
		app.config.ServerConfig = savingConfig.ServerConfig
		app.config.DockerConfig = savingConfig.DockerConfig
		app.config.KubernetesConfig = savingConfig.KubernetesConfig
		app.config.EventBusConfig = savingConfig.EventBusConfig
		app.config.APIGatewayConfig = savingConfig.APIGatewayConfig
		app.config.MgrDashboardConfig = savingConfig.MgrDashboardConfig
		app.config.Deployment = savingConfig.Deployment

	}

}

//SaveConfiguration saveConfiguration in .config file
func (app *App) SaveConfiguration() {

	configFilePath := mgrconfig.GetConfigFilePath()

	//creating config dir
	os.Mkdir(config.GetConfigDirPath(), os.FileMode.Perm(0777))

	//write default configurations
	wrireConfigJSON, _ := yaml.Marshal(mgrconfig.SavingConfig{
		Auth:               app.config.Auth,
		ServerConfig:       app.config.ServerConfig,
		DockerConfig:       app.config.DockerConfig,
		EventBusConfig:     app.config.EventBusConfig,
		APIGatewayConfig:   app.config.APIGatewayConfig,
		MgrDashboardConfig: app.config.MgrDashboardConfig,
		KubernetesConfig:   app.config.KubernetesConfig,
		Deployment:         app.config.Deployment,
	})

	//write default configurations into config file
	err := ioutil.WriteFile(configFilePath, wrireConfigJSON, 0777)
	if err != nil {
		log.Fatalf("unable to create config file %v\n", err)
	}

}

func (app *App) setupAdminUser() {

	adminUser := types.User{
		Username: app.config.Auth.Username,
		Password: app.config.Auth.Password,
	}

	dao.Add(app.db, &adminUser)

}

func (app *App) setupDeployment() {

	deployment := app.config.Deployment

	if deployment == mgrconfig.Deployment_Docker {

		dockerDeployment := docker_deployment.Deployment{
			Config: docker_deployment.Config{
				NetworkName: common.DockerNetworkID,
			},
		}

		app.deployment = dockerDeployment

	} else if deployment == mgrconfig.Deployment_Kubernetes {

		kubeDeployment := kube_deployment.Deployment{
			Config: kube_deployment.Config{
				ConfigPath: app.config.KubernetesConfig.ConfigPath,
			},
		}

		app.deployment = kubeDeployment

	} else {
		log.Fatalf("deployment setup failed : %s not support", deployment)
	}

	err := app.deployment.Init()
	if err != nil {
		log.Fatalf("deployment init failed : %v", err)
	}

}

func (app *App) setupManagerComponents() {

	deployment := app.deployment

	err := components.EventbusSetup(&app.config, deployment)
	if err != nil {
		log.Fatalf("eventbus setup failed : %v", err)
	}

	//setup messenger
	app.setupMessenger()

	//apigateway data serve
	app.setupApigatewayDdataFetchListener()

	//setup logger
	app.setupLogger()

	//apigateway
	err = components.ApigatewaySetup(&app.config, deployment)
	if err != nil {
		log.Fatalf("apigateway setup failed : %v", err)
	}

	//mgr-dashboard
	err = components.MgrDashboardSetup(&app.config)
	if err != nil {
		log.Printf("mgr-dashboard setup failed")
		log.Printf("%v", err)
	} else {
		mgrDashboardAddress := "127.0.0.1" + ":" + common.IntToStr(app.config.MgrDashboardConfig.ServerConfig.Port)
		log.Printf("quebic-faas-dashboard running on %s\n", mgrDashboardAddress)
	}

}

func (app *App) setupMessenger() {

	config := app.config
	messenger := _messenger.Messenger{AppID: config.AppID, EventBusConfig: config.EventBusConfig}
	err := messenger.WaitInit(time.Minute * 10)
	if err != nil {
		log.Fatalf("unable to connect eventbus %v\n", err)
	}

	app.messenger = messenger

}

func (app *App) setupApigatewayDdataFetchListener() {

	messenger := app.messenger

	err := messenger.Subscribe(common.EventApigatewayDataFetch, func(event _messenger.BaseEvent) {

		apigatewayData := types.ApigatewayData{}

		//resources
		var resources []types.Resource
		err := dao.GetAll(app.db, &types.Resource{}, func(k, v []byte) error {

			resource := types.Resource{}
			json.Unmarshal(v, &resource)
			resources = append(resources, resource)
			return nil
		})
		if err != nil {
			messenger.ReplyError(event, err.Error(), 500)
			return
		}

		//manager-components
		allowComponents := [1]string{
			common.ComponentEventBus,
		}
		var components []types.ManagerComponent
		err = dao.GetAll(app.db, &types.ManagerComponent{}, func(k, v []byte) error {

			component := types.ManagerComponent{}
			json.Unmarshal(v, &component)

			//check this component is allow to serve
			for _, a := range allowComponents {
				if a == component.ID {
					components = append(components, component)
				}
			}

			return nil

		})
		if err != nil {
			messenger.ReplyError(event, err.Error(), 500)
			return
		}

		//remove nil from response
		if components == nil {
			components = make([]types.ManagerComponent, 0)
		}

		if resources == nil {
			resources = make([]types.Resource, 0)
		}

		//assign data
		apigatewayData.Resources = resources
		apigatewayData.ManagerComponents = components

		messenger.ReplySuccess(event, apigatewayData, 200)

	}, common.ConsumerMgr)
	if err != nil {
		log.Fatalf("unable to subscribe internal message listen %v\n", err)
	}

}

func (app *App) setupLogger() {

	loggerUtil := logger.Logger{}
	loggerUtil.Init(app.db, app.messenger)
	loggerUtil.Listen()

	app.loggerUtil = loggerUtil

}

//GetDB get app db connection
func (app *App) GetDB() *bolt.DB {
	return app.db
}

func (app *App) setUpHTTPHandlers() {

	router := app.router
	db := app.db
	messenger := app.messenger
	loggerUtil := app.loggerUtil
	deployment := app.deployment

	httphandler.SetUpHTTPHandlers(
		app.config,
		router,
		db,
		messenger,
		loggerUtil,
		deployment)

	address := "127.0.0.1" + ":" + common.IntToStr(app.config.ServerConfig.Port)

	log.Printf("quebic-faas-manager running on %s\n", address)

	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})

	err := http.ListenAndServe(address, handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(router))
	if err != nil {
		log.Fatalf("quebic-faas-manager failed. error : %v", err)
	}

}

//GetAppConfig get app config
func (app *App) GetAppConfig() *mgrconfig.AppConfig {
	return &app.config
}
