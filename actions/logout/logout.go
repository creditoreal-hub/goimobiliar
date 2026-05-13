package logout

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/creditoreal-hub/goimobiliar/erros"
)

var ACTION = "LOGOUT"

type RunInput HandlerInput
type RunOutput HandlerOutput

func Run(input *RunInput) (*RunOutput, error) {
	handlerOutput, err := handler((*HandlerInput)(input))

	return (*RunOutput)(handlerOutput), err
}

type HandlerInput struct {
	Endpoint  string
	SessionId string `json:"SessionId,omitempty"`
}

type HandlerOutput struct {
	*RequestResponse
}

type Request struct {
	Header *RequestHeader `json:"Header,omitempty"`
	Body   *RequestBody   `json:"Body"`
}

type RequestHeader struct {
	SessionId string `json:"SessionId"`
	Action    string `json:"Action"`
}

type RequestBody struct {
}

type RequestResponse struct {
	Header *RequestResponseHeader
}

type RequestResponseHeader struct {
	SessionId string `json:"SessionId,omitempty"`
	Action    string `json:"Action,omitempty"`
	Status    string `json:"Status,omitempty"`
	Error     bool   `json:"Error,omitempty"`
}

func handler(input *HandlerInput) (*HandlerOutput, error) {
	request := Request{
		Header: &RequestHeader{
			SessionId: input.SessionId,
			Action:    ACTION,
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", input.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	r.Header.Add("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	byteBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = erros.CheckResponseError(&byteBody)
	if err != nil {
		return nil, err
	}

	var requestResponse RequestResponse
	if err := json.Unmarshal(byteBody, &requestResponse); err != nil {
		return nil, err
	}

	handlerOutput := HandlerOutput{
		RequestResponse: &requestResponse,
	}

	return &handlerOutput, nil
}
