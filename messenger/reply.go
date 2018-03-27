/*
Copyright 2018 Tharanga Nilupul Thennakoon

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package messenger

import (
	"fmt"
)

//ReplySuccess reply success response
func (messenger *Messenger) ReplySuccess(requestBaseEvent BaseEvent, replyPayload interface{}, statuscode int) error {

	_, err := messenger.publish(
		requestBaseEvent.GetRequestID(),
		replyPayload,
		nil,
		statuscode,
		emptyError,
		nil,
		nil,
		0,
		false)

	if err != nil {
		return fmt.Errorf("failed to reply success, error : %v", err)
	}

	return nil
}

//ReplyError reply error response
func (messenger *Messenger) ReplyError(requestBaseEvent BaseEvent, errMessage string, statuscode int) error {

	_, err := messenger.publish(
		requestBaseEvent.GetRequestID(),
		nil,
		nil,
		statuscode,
		errMessage,
		nil,
		nil,
		0,
		false)

	if err != nil {
		return fmt.Errorf("failed to reply error, error : %v", err)
	}

	return nil
}
