package service

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"quebic-faas/types"
	"strings"
)

const request_get = "GET"
const request_post = "POST"
const request_put = "PUT"
const request_delete = "DELETE"

//GET request
func (mgrService *MgrService) GET(path string, payload interface{}, header map[string]string) (*ResponseMessage, *types.ErrorResponse) {
	return mgrService.makeRequest(path, request_get, payload, header)
}

//POST request
func (mgrService *MgrService) POST(path string, payload interface{}, header map[string]string) (*ResponseMessage, *types.ErrorResponse) {
	return mgrService.makeRequest(path, request_post, payload, header)
}

//PUT request
func (mgrService *MgrService) PUT(path string, payload interface{}, header map[string]string) (*ResponseMessage, *types.ErrorResponse) {
	return mgrService.makeRequest(path, request_put, payload, header)
}

//DELETE request
func (mgrService *MgrService) DELETE(path string, payload interface{}, header map[string]string) (*ResponseMessage, *types.ErrorResponse) {
	return mgrService.makeRequest(path, request_delete, payload, header)
}

func (mgrService *MgrService) makeRequest(path string, method string, payload interface{}, header map[string]string) (*ResponseMessage, *types.ErrorResponse) {

	url := mgrService.prepareURL(path)

	var jsonPayload []byte
	if payload != nil {
		jsonPayload, _ = json.Marshal(payload)
	} else {
		jsonPayload = []byte("")
	}

	requestBody := strings.NewReader(string(jsonPayload))

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, makeErrorToErrorResponse(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", mgrService.Auth.AuthToken)

	if header != nil {
		for k, v := range header {
			req.Header.Add(k, v)
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, makeErrorToErrorResponse(err)
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, makeErrorToErrorResponse(err)
	}

	return &ResponseMessage{StatusCode: res.StatusCode, Data: responseBody}, nil

}
