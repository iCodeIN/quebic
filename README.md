# Quebic - FaaS Framework

Quebic is a framework for writing serverless functions to run on Dockers or Kubernetes. You can write your functions in any language. Currently quebic supports only for Java and NodeJS.

![quebic](https://github.com/quebic-source/quebic/blob/master/docs/quebic.png)

### Getting Started
#### Install Docker
 * If you have allready setup docker on your envirnment , skip this step.
 * [Install Docker guidlines](https://docs.docker.com/install/)
#### Linux Users
 * Download the binary files in [here](https://github.com/quebic-source/quebic/blob/master/bin/quebic.tar.gz).
#### Windows User
 * Clone this project then build using [golang](https://golang.org/).
 * You have to install [govendor](https://github.com/kardianos/govendor) dependency before starting to build.
 * Then you can use govendor for downloading all the required dependencies.
#### quebic-mgr
 * First thing is run **quebic-mgr [options]**
 * quebic-mgr spin-up event-bus and api-gateway.
 * We will disucuss more details about the options in future section. For not just check **quebic-mgr -h** 

#### quebic cli
 * quebic cli is an interactive commond line tool. You can easily manage your functions, routing, or any other components by using cli. Lets look at it later.
 * Sample quebic cli commond
 * **quebic function create --file function_spec_file.yml**
 * **quebic function ls**
 
### Functions
#### Java Runtime
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
#### RequestHandler<Request, Response>
 * RequestHandler is an interface which comes with quebic-runtime-java library. You can add your logic inside it's handle() method.

