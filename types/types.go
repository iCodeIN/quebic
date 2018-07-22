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

package types

import (
	"quebic-faas/common"
	"reflect"
)

//Entity is used as Base for every models ######################################
type Entity interface {
	GetReflectObject() reflect.Value
	GetID() string
	SetID(id string)
	SetModifiedAt()
}

//User model ######################################
type User struct {
	Username   string `json:"username"`
	Firstname  string `json:"firstname"`
	Password   string `json:"password"`
	CreatedAt  string `json:"createdAt" yaml:"createdAt"`   //created time
	ModifiedAt string `json:"modifiedAt" yaml:"modifiedAt"` //modified time
}

//GetReflectObject get Reflect Object
func (o *User) GetReflectObject() reflect.Value {
	return reflect.ValueOf(o)
}

//GetID get ID
func (o *User) GetID() string {
	return o.Username
}

//SetID get ID
func (o *User) SetID(id string) {
	o.Username = id
}

//SetModifiedAt set modified date
func (o *User) SetModifiedAt() {
	o.ModifiedAt = common.CurrentTime()
}

//Event model ######################################
//ID => <prefix>.Group.Name
type Event struct {
	ID         string `json:"id"`
	Group      string `json:"group"`
	Name       string `json:"name"`
	CreatedAt  string `json:"createdAt" yaml:"createdAt"`   //created time
	ModifiedAt string `json:"modifiedAt" yaml:"modifiedAt"` //modified time
}

//GetReflectObject get Reflect Object
func (o *Event) GetReflectObject() reflect.Value {
	return reflect.ValueOf(o)
}

//GetID get ID
func (o *Event) GetID() string {
	return o.ID
}

//SetID get ID
func (o *Event) SetID(id string) {
	o.ID = id
}

//SetModifiedAt set modified date
func (o *Event) SetModifiedAt() {
	o.ModifiedAt = common.CurrentTime()
}

//Resource model ######################################
//ID => URL:RequestMethod
type Resource struct {
	ID                    string                   `json:"id"`
	Name                  string                   `json:"name" yaml:"name"`
	RequestMethod         string                   `json:"requestMethod" yaml:"requestMethod"`
	URL                   string                   `json:"url" yaml:"url"`
	Async                 bool                     `json:"async" yaml:"async"`
	RequestTimeout        int                      `json:"requestTimeout" yaml:"requestTimeout"` //in second
	SuccessResponseStatus int                      `json:"successResponseStatus" yaml:"successResponseStatus"`
	Event                 string                   `json:"event" yaml:"event"`
	RequestMapping        []RequestMappingTemplate `json:"requestMapping" yaml:"requestMapping"`
	HeaderMapping         []HeaderMappingTemplate  `json:"headerMapping" yaml:"headerMapping"`
	HeadersToPass         []string                 `json:"headersToPass" yaml:"headersToPass"` //header list pass to endpoint
	CreatedAt             string                   `json:"createdAt" yaml:"createdAt"`         //created time
	ModifiedAt            string                   `json:"modifiedAt" yaml:"modifiedAt"`       //modified time
}

//GetReflectObject get Reflect Object
func (o *Resource) GetReflectObject() reflect.Value {
	return reflect.ValueOf(o)
}

//GetID get ID
func (o *Resource) GetID() string {
	return o.ID
}

//SetID get ID
func (o *Resource) SetID(id string) {
	o.ID = id
}

//SetDefault set default values
func (o *Resource) SetDefault() {
	o.RequestTimeout = 1000 * 10 //10 seconds
	o.SuccessResponseStatus = 200
}

//SetModifiedAt set modified date
func (o *Resource) SetModifiedAt() {
	o.ModifiedAt = common.CurrentTime()
}

//Function model ######################################
// Source : code / artifact stored file location
// FunctionPath :
// 		Runtime java : class path of handler
//		Runtime node : myHandler
// FunctionFile :
//      Runtime jsvs : const value eg : function.jar
//      Runtime node : const value eg : handler.js
//      Runtime node package : user defined
// Route : function invoker route id. nor required
type Function struct {
	Name                 string                `json:"name" yaml:"name"`
	Version              string                `json:"version" yaml:"version"` //current version
	Versions             []string              `json:"versions" yaml:"versions"`
	DockerImageID        string                `json:"dockerImageID" yaml:"dockerImageID"`
	Source               string                `json:"source" yaml:"source"`
	Handler              string                `json:"handler" yaml:"handler"`
	HandlerPath          string                `json:"handlerPath" yaml:"handlerPath"`
	HandlerFile          string                `json:"handlerFile" yaml:"handlerFile"`
	Runtime              string                `json:"runtime" yaml:"runtime"`
	Events               []string              `json:"events" yaml:"events"`
	Replicas             int                   `json:"replicas" yaml:"replicas"`
	EnvironmentVariables []EnvironmentVariable `json:"env" yaml:"env"`
	SecretKey            string                `json:"secretKey"`
	Route                string                `json:"route"`
	Log                  EntityLog             `json:"log"`
	ModifiedAt           string                `json:"modifiedAt" yaml:"modifiedAt"` //modified time
	Status               string                `json:"status" yaml:"status"`
}

//GetReflectObject get Reflect Object
func (o *Function) GetReflectObject() reflect.Value {
	return reflect.ValueOf(o)
}

//GetID get ID
func (o *Function) GetID() string {
	return o.Name
}

//SetID get ID
func (o *Function) SetID(id string) {
	o.Name = id
}

//SetModifiedAt set modified date
func (o *Function) SetModifiedAt() {
	o.ModifiedAt = common.CurrentTime()
}

//EnvironmentVariable environmentVariable
type EnvironmentVariable struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

//ManagerComponent components handle by manager ######################################
//ID + AccessKey => JWT => store in dockerimage => later that JWT token will use to authenticate resource access
type ManagerComponent struct {
	ID                string     `json:"id"`
	DockerImageID     string     `json:"dockerImageID"`
	DockerContainerID string     `json:"dockerContainerID"`
	AccessKey         string     `json:"accessKey"`
	Deployment        Deployment `json:"deployment"`
	Log               EntityLog  `json:"log"`
	CreatedAt         string     `json:"createdAt" yaml:"createdAt"`   //created time
	ModifiedAt        string     `json:"modifiedAt" yaml:"modifiedAt"` //modified time
}

//GetReflectObject get Reflect Object
func (o *ManagerComponent) GetReflectObject() reflect.Value {
	return reflect.ValueOf(o)
}

//GetID get ID
func (o *ManagerComponent) GetID() string {
	return o.ID
}

//SetID get ID
func (o *ManagerComponent) SetID(id string) {
	o.ID = id
}

//SetModifiedAt set modified date
func (o *ManagerComponent) SetModifiedAt() {
	o.ModifiedAt = common.CurrentTime()
}

//Deployment deployment details
type Deployment struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

//RequestMappingTemplate model ######################################
type RequestMappingTemplate struct {
	EventAttribute   string `json:"eventAttribute" yaml:"eventAttribute"`
	RequestAttribute string `json:"requestAttribute" yaml:"requestAttribute"`
}

//HeaderMappingTemplate model ######################################
type HeaderMappingTemplate struct {
	EventAttribute  string `json:"eventAttribute" yaml:"eventAttribute"`
	HeaderAttribute string `json:"headerAttribute" yaml:"headerAttribute"`
}

//EntityLog log entity is used to describe states of each entities ######################################
type EntityLog struct {
	Time    string `json:"time"`
	State   string `json:"state"`
	Message string `json:"message"`
}

//ApigatewayData used to get data from manager to apigateway
type ApigatewayData struct {
	ManagerComponents []ManagerComponent `json:"managerComponents"`
	Resources         []Resource         `json:"resources"`
}

//RequestTracker requestTracker
type RequestTracker struct {
	RequestID   string                 `json:"requestID"`
	Source      string                 `json:"source"`      //created by
	Response    RequestTrackerResponse `json:"response"`    //response set by function
	CreatedAt   string                 `json:"createdAt"`   //created time
	CompletedAt string                 `json:"completedAt"` //completed time
	Logs        []Log                  `json:"logs"`        //logs created by function
}

//GetReflectObject get Reflect Object
func (o *RequestTracker) GetReflectObject() reflect.Value {
	return reflect.ValueOf(o)
}

//GetID get ID
func (o *RequestTracker) GetID() string {
	return o.RequestID
}

//SetID get ID
func (o *RequestTracker) SetID(id string) {
	o.RequestID = id
}

//SetModifiedAt set modified date
func (o *RequestTracker) SetModifiedAt() {
}

//RequestTrackerResponse requestTrackerResponse
type RequestTrackerResponse struct {
	Status  int         `json:"status"`
	Message interface{} `json:"message"`
}

//Log log
type Log struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Source  string `json:"source"` //executed by function-container-id or app-id
	Time    string `json:"time"`   //executed time
}

//RequestTrackerMessage use to communtication
type RequestTrackerMessage struct {
	RequestID string                 `json:"requestID"`
	Response  RequestTrackerResponse `json:"response"`
	Log       Log                    `json:"log"`
	Completed bool                   `json:"completed"`
}
