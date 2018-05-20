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

package httphandler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"quebic-faas/common"
	"quebic-faas/quebic-faas-mgr/config"
	"quebic-faas/quebic-faas-mgr/dao"
	"quebic-faas/quebic-faas-mgr/functionutil"
	"quebic-faas/quebic-faas-mgr/functionutil/function_create"
	"quebic-faas/types"
	"strings"

	dep "quebic-faas/quebic-faas-mgr/deployment"

	bolt "github.com/coreos/bbolt"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

const initialVersion = "0.1.0"

//FunctionHandler handler
func (httphandler *Httphandler) FunctionHandler(router *mux.Router) {

	db := httphandler.db
	appConfig := httphandler.config
	deployment := httphandler.deployment

	router.HandleFunc("/functions", func(w http.ResponseWriter, r *http.Request) {

		qID := r.FormValue("id")
		if qID == "" {
			getAllFunctions(w, r, db, &types.Function{})
			return
		}

		function := &types.Function{}
		function.Name = qID
		getByID(w, r, db, function)

	}).Methods("GET")

	router.HandleFunc("/functions/{name}", func(w http.ResponseWriter, r *http.Request) {
		getByID(w, r, db, processRequestParmForID(r, "name", &types.Function{}))
	}).Methods("GET")

	router.HandleFunc("/functions", func(w http.ResponseWriter, r *http.Request) {

		functionDTO := &types.FunctionDTO{}
		err := processFunctionSaveReqest(r, functionDTO)
		if err != nil {
			makeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		saveFunctionDTO(w, r, db, functionDTO, appConfig, deployment, true)

	}).Methods("POST")

	router.HandleFunc("/functions", func(w http.ResponseWriter, r *http.Request) {

		functionDTO := &types.FunctionDTO{}
		err := processFunctionSaveReqest(r, functionDTO)
		if err != nil {
			makeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		saveFunctionDTO(w, r, db, functionDTO, appConfig, deployment, false)

	}).Methods("PUT")

	router.HandleFunc("/functions/deploy", func(w http.ResponseWriter, r *http.Request) {

		function := &types.Function{}
		err := processRequest(r, function)
		if err != nil {
			makeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		trimStringFieldsFunction(function)

		requestVersion := function.Version

		errors := validationFunctionContainer(db, function)
		if errors != nil {
			status := http.StatusBadRequest
			writeResponse(w, types.ErrorResponse{Cause: "validation-failed", Message: errors, Status: status}, status)
			return
		}

		if !checkVersionIsExists(*function, requestVersion) {
			status := http.StatusNotAcceptable
			writeResponse(w, types.ErrorResponse{Cause: "not-acceptable", Message: []string{"There is not any build match for requested version"}, Status: status}, status)
			return
		}

		function.Version = requestVersion

		err = dao.Update(db, function)
		if err != nil {
			makeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		_, err = functionutil.FunctionDeploy(
			appConfig,
			deployment,
			function)
		if err != nil {
			makeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		writeResponse(w, function, 200)

	}).Methods("POST")

	router.HandleFunc("/functions", func(w http.ResponseWriter, r *http.Request) {

		function := &types.Function{}
		err := processRequest(r, function)
		if err != nil {
			makeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		trimStringFieldsFunction(function)

		errors := validationFunctionContainer(db, function)
		if errors != nil {
			status := http.StatusBadRequest
			writeResponse(w, types.ErrorResponse{Cause: "validation-failed", Message: errors, Status: status}, status)
			return
		}

		err = dao.Delete(db, function)
		if err != nil {
			makeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		//stop container
		err = functionutil.StopFunction(appConfig, deployment, function)
		if err != nil {
			makeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		writeResponse(w, function, 200)

	}).Methods("DELETE")

	router.HandleFunc("/function_containers/logs", func(w http.ResponseWriter, r *http.Request) {

		if deployment.DeploymentType() != config.Deployment_Docker {
			makeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("funtion-logs access only allow for docker deployment mode"))
			return
		}

		logDTO := &types.FunctionContainerLogDTO{}
		err := processRequest(r, logDTO)
		if err != nil {
			makeErrorResponse(w, http.StatusBadRequest, err)
			return
		}

		errors := validationFunctionContainer(db, &logDTO.Function)
		if errors != nil {
			status := http.StatusBadRequest
			writeResponse(w, types.ErrorResponse{Cause: "validation-failed", Message: errors, Status: status}, status)
			return
		}

		options := types.FunctionContainerLogOptions{
			Details:    logDTO.Options.Details,
			Follow:     logDTO.Options.Follow,
			ShowStderr: logDTO.Options.ShowStderr,
			ShowStdout: logDTO.Options.ShowStdout,
			Since:      logDTO.Options.Since,
			Tail:       logDTO.Options.Tail,
			Timestamps: logDTO.Options.Timestamps,
			Until:      logDTO.Options.Until,
		}

		functionService := functionutil.GetServiceID(logDTO.Function.Name)
		logs, err := deployment.LogsByName(functionService, options)
		if err != nil {
			makeErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

		writeResponse(w, logs, 200)

	}).Methods("POST")

}

func processFunctionSaveReqest(r *http.Request, functionDTO *types.FunctionDTO) error {

	fieldSpec := common.FunctionSaveField_SPEC
	fieldSource := common.FunctionSaveField_SOURCE

	const MEMORY = 5 * 1024 * 1024 //5mb
	r.ParseMultipartForm(MEMORY)

	//Spec Data
	specJSON := r.Form.Get(fieldSpec)
	if specJSON == "" {
		return fmt.Errorf("unable to find %s field in request", fieldSpec)
	}

	err := json.Unmarshal([]byte(specJSON), functionDTO)
	if err != nil {
		return fmt.Errorf("%s data not in correct format", fieldSpec)
	}

	//Source
	sourceFile, sourceFileHandler, err := r.FormFile(fieldSource)
	if err != nil {
		return fmt.Errorf("unable to load %s file in request", fieldSource)
	}

	functionDTO.SourceFile.File = sourceFile
	functionDTO.SourceFile.FileHeader = sourceFileHandler

	return nil
}

func getAllFunctions(w http.ResponseWriter, r *http.Request, db *bolt.DB, entity types.Entity) {

	var functions []types.Function
	err := dao.GetAll(db, entity, func(k, v []byte) error {

		function := types.Function{}
		json.Unmarshal(v, &function)
		functions = append(functions, function)
		return nil
	})

	if err != nil {
		makeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if functions == nil {
		var emptyStr [0]string
		writeResponse(w, emptyStr, http.StatusOK)
	} else {
		writeResponse(w, functions, http.StatusOK)
	}

}

func getFunctionByID(db *bolt.DB, function *types.Function) error {

	err := dao.GetByID(db, function, func(savedObj []byte) error {

		if savedObj == nil {
			return makeError("function not found", nil)
		}

		json.Unmarshal(savedObj, function)

		return nil
	})

	if err != nil {
		return err
	}

	return nil

}

func saveFunctionDTO(
	w http.ResponseWriter,
	r *http.Request,
	db *bolt.DB,
	functionDTO *types.FunctionDTO,
	appConfig config.AppConfig,
	deployment dep.Deployment,
	isCreate bool) {

	function := &functionDTO.Function
	route := &functionDTO.Route
	fillRouteDataFromFunction(function, route)

	//trim
	trimStringFieldsFunction(function)
	trimStringFieldsResource(route)

	errors := validateFunctionDTO(db, functionDTO, isCreate)
	if len(errors) > 0 {
		status := http.StatusBadRequest
		writeResponse(w, types.ErrorResponse{Cause: "validation-failed", Message: errors, Status: status}, status)
		return
	}

	err := saveFunction(db, functionDTO, appConfig, isCreate)
	if err != nil {
		makeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	_, err = functionutil.FunctionDeploy(
		appConfig,
		deployment,
		function)
	if err != nil {
		makeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	err = saveRoute(db, route, appConfig, deployment, isCreate)
	if err != nil {
		makeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	writeResponse(w, functionDTO, 200)

}

func fillRouteDataFromFunction(function *types.Function, route *types.Resource) {
	//fill function's data
	route.Name = function.Name
	route.Event = prepareFunctionEvent(function.GetID())
}

func saveFunction(
	db *bolt.DB,
	functionDTO *types.FunctionDTO,
	appConfig config.AppConfig,
	isCreate bool) error {

	function := &functionDTO.Function

	err := preProcessFunction(function)
	if err != nil {
		return err
	}

	if function.GetID() == "" {
		return fmt.Errorf("unable find id field")
	}

	if isCreate {
		err = dao.Add(db, function)
		if err != nil {
			return err
		}
	} else {
		err = dao.Update(db, function)
		if err != nil {
			return err
		}
	}

	return postProcessFunction(db, appConfig.DockerConfig, functionDTO)

}

func saveRoute(db *bolt.DB, route *types.Resource, appConfig config.AppConfig, deployment dep.Deployment, isCreate bool) error {

	err := preProcessResource(db, route)
	if err != nil {
		return err
	}

	if route.GetID() == "" {
		return nil
	}

	/*if isCreate {
		err = dao.Add(db, route)
		if err != nil {
			return err
		}
	} else {
		err = dao.Update(db, route)
		if err != nil {
			return err
		}
	}*/

	route.SetModifiedAt()
	err = dao.Save(db, route)
	if err != nil {
		return err
	}

	//re-start apigateway
	restartAPIGateway(appConfig, deployment)

	return nil

}

func validateFunctionDTO(db *bolt.DB, functionDTO *types.FunctionDTO, isCreate bool) map[string][]string {

	var errors = make(map[string][]string)

	functionErrors := validationFunction(db, &functionDTO.Function, functionDTO.SourceFile, isCreate)
	if functionErrors != nil {
		errors["function-validation-errors"] = functionErrors
	}

	if functionDTO.Route.URL != "" {
		routeErrors := validationRoute(db, &functionDTO.Route, true, isCreate)

		//ignore id validation
		if !(len(routeErrors) == 1 && "route is not found" == routeErrors[0]) {
			if routeErrors != nil {
				errors["route-validation-errors"] = routeErrors
			}
		}

	}

	functionDTO.Function.Route = functionDTO.Route.GetID()

	return errors

}

func trimStringFieldsFunction(function *types.Function) {
	function.Name = Trim(function.Name)
}

func validationFunction(db *bolt.DB, function *types.Function, functionArtifactFile types.FunctionSourceFile, isCreate bool) []string {

	var errors []string

	if function.Name == "" {
		errors = append(errors, "name field should not be empty")
	}

	if strings.Contains(function.Name, " ") {
		errors = append(errors, "name field not allow to contain spaces")
	}

	if function.Runtime == "" {
		errors = append(errors, "runtime field should not be empty")
	}

	if function.Handler == "" {
		errors = append(errors, "handler field should not be empty")
	}

	if function.Runtime != "" {

		if !common.RuntimeValidate(common.Runtime(function.Runtime)) {
			errors = append(errors, "runtime not match")
		} else {

			if function.Handler != "" {

				err := prepareHandler(function, functionArtifactFile)
				if err != nil {
					errors = append(errors, err.Error())
				}

			}

		}

	}

	if function.Replicas < 0 {
		errors = append(errors, "replicas value is invalide")
	}

	if function.Events != nil {

		for i, eventID := range function.Events {

			event, eventErrors := createEventFromEventID(db, eventID)
			if eventErrors != nil {
				errors = append(errors, eventErrors...)
			} else {
				function.Events[i] = event.GetID()
			}

		}
	}

	if isCreate {

		if checkFunctionISAlreadyExists(db, function) {
			errors = append(errors, "function is already exists")
		}

	} else {

		if !checkFunctionISAlreadyExists(db, function) {
			errors = append(errors, "function not found")
		}

	}

	return errors

}

func validationFunctionContainer(db *bolt.DB, function *types.Function) []string {

	var errors []string

	if function.Name == "" {
		errors = append(errors, "function name should not be empty")
	}

	if function.Name != "" {
		err := getFunctionByID(db, function)
		if err != nil {
			errors = append(errors, err.Error())
		}
	}

	return errors

}

func checkFunctionISAlreadyExists(db *bolt.DB, function *types.Function) bool {

	newVersion := function.Version

	found := false

	_ = dao.GetByID(db, function, func(savedObj []byte) error {

		if savedObj != nil {

			found = true

			//Version setup
			savedFunction := &types.Function{}
			json.Unmarshal(savedObj, savedFunction)

			function.Versions = savedFunction.Versions

			if newVersion == "" || newVersion == "latest" {
				function.Version = savedFunction.Version
			} else {
				if !checkVersionIsExists(*savedFunction, newVersion) {
					function.Versions = append(function.Versions, function.Version)
				}
			}
			//Version setup

		} else {

			//Version setup
			if newVersion == "" || newVersion == "latest" {
				function.Version = initialVersion
			}
			function.Versions = append(function.Versions, function.Version)
			//Version setup

		}

		return nil
	})

	return found
}

func preProcessFunction(
	function *types.Function) error {

	//SecretKey
	secretKeyUUID, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("unable to assign function secretKey %v", err)
	}

	function.SecretKey = secretKeyUUID.String()

	//Replicas
	if function.Replicas <= 0 {
		function.Replicas = 1
	}

	//Events
	function.Events = append(function.Events, prepareFunctionEvent(function.GetID()))

	return nil
}

func prepareFunctionEvent(functionID string) string {
	return common.EventPrefixFunction + common.EventJOIN + functionID
}

func prepareHandler(function *types.Function, functionArtifactFile types.FunctionSourceFile) error {
	functionRuntime := function.Runtime
	functionArtifactFilename := functionArtifactFile.FileHeader.Filename

	if functionRuntime == common.RuntimeJava {

		if filepath.Ext(functionArtifactFilename) == ".jar" {
			function.HandlerFile = function_create.GetDockerFunctionJAR()
			function.HandlerPath = function.Handler
			return nil
		}

	} else if functionRuntime == common.RuntimeNodeJS {

		// Ex : handler = index.myHandler
		//handler = {handlerFile}.{handlerPath}
		h := strings.Split(function.Handler, ".")

		if len(h) < 2 {
			return makeError("handler is invalide. unable to found module", nil)
		}

		handlerFile := h[0] + ".js"
		handlerPath := h[1]

		function.HandlerPath = handlerPath

		//node package
		if filepath.Ext(functionArtifactFilename) == ".tar" || filepath.Ext(functionArtifactFilename) == ".gz" {

			if handlerFile == "" {
				return fmt.Errorf("handler cannot be empty for node package")
			}

			function.HandlerFile = function_create.GetDockerFunctionJSPackage(handlerFile)

			return nil

		} else if filepath.Ext(functionArtifactFilename) == ".js" {

			function.HandlerFile = function_create.GetDockerFunctionJS()
			return nil

		}

	}

	return makeError("artifact file type invalide for runtime", nil)

}

func postProcessFunction(
	db *bolt.DB,
	dockerConfig config.DockerConfig,
	functionDTO *types.FunctionDTO) error {

	function := &functionDTO.Function

	entityLog := types.EntityLog{State: common.LogStateSaved}
	dao.AddFunctionLog(db, function, entityLog)

	authConfig, err := dockerConfig.GetDockerAuthConfig()
	if err != nil {
		log.Print(err)
		return err
	}

	dockerImageID, err := functionutil.FunctionCreate(authConfig, *functionDTO)

	if err != nil {
		entityLog = types.EntityLog{State: common.LogStateDockerImageCreatingFailed, Message: err.Error()}
		dao.AddFunctionLog(db, function, entityLog)
		return err
	}

	dao.AddFunctionDockerImageID(db, function, dockerImageID)
	entityLog = types.EntityLog{State: common.LogStateDockerImageCreated, Message: dockerImageID}
	dao.AddFunctionLog(db, function, entityLog)

	function.DockerImageID = dockerImageID

	return nil

}

func checkVersionIsExists(function types.Function, version string) bool {

	for _, v := range function.Versions {

		if v == version {
			return true
		}

	}

	return false
}
