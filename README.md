# Quebic - FaaS Framework

Quebic is a framework for writing serverless functions to run on Kubernetes. Currently it supports for Python, Java and NodeJS programming models to write functions.
* [Wiki](https://github.com/quebic-source/quebic/wiki)
* [Demo](https://youtu.be/wxydEo5-wGo)
* [Example Project](https://github.com/quebic-source/quebic-sample-project)
* [Event-driven Microservices with Quebic](https://hackernoon.com/event-driven-microservices-with-quebic-f65f99a5b25a)

![quebic](https://github.com/quebic-source/quebic/blob/master/docs/quebic.png)

## Contents
* [Getting Started](#getting-started)
* [Functions](#functions)
* [Routing](#routing)
* [Asynchronous invocation from API Gateway](#async)
* [Logs](#logs)
* [Configurations](#configurations)
* [Example Project](https://github.com/quebic-source/quebic-sample-project)
* [Consultants](#consultants)
* [Releases](https://github.com/quebic-source/quebic/releases)

## <a name="getting-started"></a>Getting Started

#### Install Docker
 * If you have already setup docker on your environment ,skip this step.
 * [Install Docker](https://docs.docker.com/install/)

#### Getting Binaries

###### For Linux Users
 * Download binaries from [here](https://github.com/quebic-source/quebic/releases). Save and extract it into preferred location.
 * After extract, you can see quebic-mgr and quebic cli inside that dir. 
 
###### For Windows Users
 * [Install golang into your environment](https://golang.org/doc/install). 
 * Get [govendor](https://github.com/kardianos/govendor) tool. 
 * Run **govendor fetch**. This will download all the required dependencies for quebic.
 * Run for build quebic-mgr **go install quebic-faas/quebic-faas-mgr**
 * Run for build quebic cli **go install quebic-faas/quebic-faas-cli**
 * Congrats !!! Now you can find your binaries from $GOPATH/bin dir.

#### Run quebic-manager
 * Jump into quebic binaries location. Then run this **quebic-mgr**
 * You can use quebic cli or quebic-mgr-dashboard ui to communicate with quebic-manager.
 * By default quebic-mgr-dashboard ui is running [localhost:8000](http://localhost:8000)
 
## <a name="functions"></a>Functions

#### Python Runtime
##### [Example](https://github.com/quebic-source/quebic-sample-project)
##### Programming Model
###### RequestHandler
 * Write your logic inside the handler.
```python
def handler(payload, context, callback):
    
    def m_callback(r):
        callback(None, 'success', 201)

    def m_error_callback(err):
        callback(None, 'failed', 403)

    context.messenger()
      .publish(
         'test.Receiver', 
         {'id': payload}, 
         m_callback, 
         m_error_callback
      )
```

###### Context
 * Context have these methods.
```python
context.base_event # return event details comes into this function
context.messenger() # return messenger instance
context.logger() # return logger instance
```

###### CallBack
* CallBack provides way to reply.
```python
callback(); # reply 200 status code with empty data
callback(null, "success")  # reply 200 status code with data
callback(null, "success", 201)  # reply 201 status code with data
callback(error) # reply 500 status code with with error-data
callback(error, null, 401) # reply 401 status code with with error-data 
```

###### Messenger
* Messenger provides way to publish events.
```python
context.messenger().publish(
    "users.UserValidate", # event id 
    user, # payload
    callback, # success callback(result)
    error_callback, # err callback(err) 
    1000 * 8 # timeout (ms)
)

```
###### Logger
* Logger provides way to attach logs for particular request context. We will discuss more about this logger in later section.
```python
context.logger().info("log info")
context.logger().error("log error")
context.logger().warn("log warn")
```

##### Deployment Spec
 * Deployment .yml spec file by describing how you want to deploy your functions into quebic.
 * Package your whole nodejs project dir into .tar file. Then set your .tar file location into source field in deployment spec.
 * If your handler is just a single python file. Then just set your .py file location into source field. No need to package it. Then handler field will be like this *index.handler*. here handler is the function
 * runtime will be python_2.7 or python_3.6
 ```yml
  function:
    name: hello-function # function name 
    source: /functions/hello-function.tar # tar package location
    handler: index.handler # request handler 
    runtime: python_2.7 # function runtime
    replicas: 2 # replicas count
    events: # function going to listen these events
      - users.UserValidate
    ...
 ```

#### Java Runtime
##### [Example](https://github.com/quebic-source/quebic-sample-project/tree/master/java-example)
##### Programming Model
###### RequestHandler<Request, Response>
 * RequestHandler is an interface which comes with [quebic-runtime-java library](https://github.com/quebic-source/quebic-runtime-java). You can add your logic inside it's handle() method.
 * The Request Type and Response Type can be any Primitive data type or Object.
```java
public class HelloFunction implements RequestHandler<Request, Response>{
 
 public void handle(Request request, CallBack<Response> callback, Context context) {
	 callback.success(new Response(1, "reply success"));
 }
 
}
```

###### Context
 * Context have these methods.
```java
BaseEvent baseEvent(); // return event details comes into this function
Messenger messenger(); // return messenger instance
Logger logger(); // return logger instance
```

###### CallBack
* CallBack provides way to reply.
```java
callback.success(); //reply 200 status code with empty data
callback.success("reply success"); //reply 200 status code with data
callback.success(201, "reply success"); //reply 201 status code with data

callBack.failure("Error occurred"); //reply 500 status code with err-data
callBack.failure(401, "Error occurred"); //reply 401 status code with error-data
```

###### Messenger
* Messenger provides way to publish events.
* void publish(String eventID, Object eventPayload, MessageHandler successHandler,ErrorHandler errorHandler,int timeout)throws MessengerException;
```java
context.messenger().publish("users.UserValidate", user, s->{
  user.setId(UUID.randomUUID().toString());
  callBack.success(201, user);
}, e->{
  callBack.failure(e.statuscode(), e.error());
}, 1000 * 8);

```
###### Logger
* Logger provides way to attach logs for particular request context. We will discuss more about this logger in later section.
```java
context.logger().info("log info");
context.logger().error("log error");
context.logger().warn("log warn");
```

##### Create .jar artifact
 * Create new maven project.
 * Add this dependency and repository into .pom file.
 ```xml
<dependency>
    <groupId>com.quebic.faas.runtime</groupId>
    <artifactId>quebic-faas-runtime-java</artifactId>
    <version>0.0.1-SNAPSHOT</version>
</dependency>

<repositories>
    <repository>
     <id>quebic-runtime-java-mvn-repo</id>
     <url>https://raw.github.com/quebic-source/quebic-runtime-java/mvn-repo/</url>
     <snapshots>
      <enabled>true</enabled>
      <updatePolicy>always</updatePolicy>
     </snapshots>
    </repository>
</repositories>
```
 * Run **mvn clean package**
 
##### Deployment Spec
 * Deployment .yml spec file by describing how you want to deploy your functions into quebic. This is code snippet for deployment spec
 ```yml
function:
  name: hello-function # function name 
  source: /functions/hello-function.jar # jar artifact location
  handler: com.quebicfaas.examples.HelloFunction # request handler java class
  runtime: java # function runtime
  events: # function going to listen these events
    - users.UserCreate
    - users.UserUpdate

route:
  requestMethod: POST
  url: /users
  requestMapping:
    - eventAttribute: eID
      requestAttribute: id
    ...
 ```
#### NodeJS Runtime
##### [Example](https://github.com/quebic-source/quebic-sample-project/tree/master/nodejs-example)
##### Programming Model
###### RequestHandler
 * Write your logic inside the handler.
```javascript
exports.validationHandler = function(payload, context, callback){
    
    if(validateUser(payload)){
        
        callback(null, true, 200);

    }else{

        callback(new Error('Not a valid e-mail address'), null, 400);

    }

}
```
###### CallBack
* CallBack provides way to reply.
```javascript
callback(); //reply 200 status code with empty data
callback(null, "success");  //reply 200 status code with data
callback(null, "success", 201);  //reply 201 status code with data
callback(error); //reply 500 status code with with error-data
callback(error, null, 401); //reply 401 status code with with error-data 
```
##### Deployment Spec
 * Deployment .yml spec file by describing how you want to deploy your functions into quebic.
 * Package your whole nodejs project dir into .tar file. Then set your .tar file location into source field in deployment spec.
 * If the handler file app.js then handler file needs to set like this *app.helloHandler*
 * If you are working on single javascript file. Then just set your .js file location into source field. No need to package.
 ```yml
  function:
    name: hello-function # function name 
    source: /functions/hello-function.tar # tar package location
    handler: app.helloHandler # request handler 
    runtime: nodejs # function runtime
    replicas: 2 # replicas count
    events: # function going to listen these events
      - users.UserValidate
    ...
 ```

#### Manage your functions with quebic cli
##### Create function
* quebic function create --file [deployment spec file]
	
##### Update function
* quebic function update --file [deployment spec file]

##### Upgrade / Downgrade function
* quebic function deploy --name [function name] --version [version]

##### Delete function
* quebic function delete --name [function name]
	
##### List all functions
* quebic function ls

##### Inspect function details
* quebic function inspect --name [function name]

## <a name="routing"></a>Routing
 * You can create routing endpoint to fire events from apigateway.
##### Routing Spec
 * Routing .yml spec is used to describe how it behave when invoke it.
```yml
name: users_route # route name just for identify
requestMethod: POST 
url: /users
async: true # enable asynchronous invocation
successResponseStatus: 201 # default response http status code
event: users.UserCreate # event going to send
requestMapping:
  - eventAttribute: eID # attribute name which function going to access in event's payload
    requestAttribute: id # attribute name which come in http request
  - eventAttribute: eName
    requestAttribute: name
headerMapping:
  - eventAttribute: auth # attribute name which function going to access in event's payload
    headerAttribute: x-token # attribute name which come in http header
headersToPass: # headers going to pass with event
  - Authorization
  - Private-Token
```
#### Manage Routes with quebic cli
##### Create Route
* quebic route create --file [route spec file]
	
##### Update Route
* quebic route update --file [route spec file]
	
##### List all Routes
* quebic route ls

##### Inspect Route details
* quebic route inspect --name [route name]


## <a name="async"></a> Asynchronous invocation from API Gateway
 * Quebic provides way to invoke function Asynchronous way from API Gateway.
 * After client send his request through the apigateway, He immediately gets a reference id (request-id) to track the request.
 * Then client can check the request by using that request-id from request-tracker endpoint of API Gateway, If function already completed the task  client will get the result of request, otherwise he will get request-still-processing message.
 ```
 <api-gateway>/request-tracker/{request-id}
 ```

## <a name="logs"></a>Logs
 * Quebic provides way to access function-container's native logs by using quebic cli.
 * **quebic function logs --name [function name]**
 * Instead of accessing native logs quebic also provides way to attach logs for particular request context. 
```java
context.logger().info("log info");
```
 * You can inspect these logs by using cli 
 * **quebic request-tracker logs --request-id [request id]**
 
 
 ## <a name="configurations"></a>Configurations
 #### Quebic manager configurations
 * Quebic manager config file is located at $HOME/.quebic-faas/manager-config.yml
 * Also you can pass arguments to the quebic manager in runtime.
 * Run **quebic-mgr -h** to list down all available commands. 
 
 #### Quebic CLI configurations
 * Quebic cli config file is located at $HOME/.quebic-faas/cli-config.yml
 * Also you can pass arguments to the quebic cli in runtime.
 * Run **quebic -h** to list down all available commands. 
 
 ##  <a name="consultants"></a>Authors
 * Tharanga Thennakoon - tharanganilupul@gmail.com 
 * [Linkedin](https://lk.linkedin.com/in/tharanga-thennakoon)

 ## License
 * This project is licensed under the Apache Licensed V2

