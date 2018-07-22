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

package common

//DockerFileContent_Java java docker file
const DockerFileContent_Java = "FROM quebicdocker/quebic-faas-container-java:1.0.0\nADD function.jar /app/function.jar\nARG access_key\nENV access_key $access_key\n"

//DockerFileContent_Java java docker file
const DockerFileContent_NodeJS = "FROM quebicdocker/quebic-faas-container-nodejs:1.0.0\nADD function_handler.tar /app/function_handler/\nARG access_key\nENV access_key $access_key\n"

//DockerFileContent_Python_2_7 docker file
const DockerFileContent_Python_2_7 = "FROM quebicdocker/quebic-faas-container-python-2_7:0.1.0\nADD function_handler.tar /app/function_handler/\nRUN pwd\n\nRUN ls\n\nRUN ls /app/function_handler/\n\nARG access_key\nENV access_key $access_key\n"

//DockerFileContent_Python_3_6 docker file
const DockerFileContent_Python_3_6 = "FROM quebicdocker/quebic-faas-container-python-3_6:0.1.0\nADD function_handler.tar /app/function_handler/\nARG access_key\nENV access_key $access_key\n"
