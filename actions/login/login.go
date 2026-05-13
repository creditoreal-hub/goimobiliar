package login

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/creditoreal-hub/goimobiliar/erros"
)

var ACTION = "LOGIN"

type ActionInput struct {
	ImobId   *string `json:"IMOB_ID,omitempty"`   // Identificação da administradora.
	UserId   *string `json:"USER_ID,omitempty"`   // Identificação do usuário.
	UserPass *string `json:"USER_PASS,omitempty"` // Senha do usário. A senha é criptografada pelo pacote.
}

type RunInput HandlerInput
type RunOutput HandlerOutput

func Run(input *RunInput) (*RunOutput, error) {
	handlerOutput, err := handler(&HandlerInput{
		Endpoint:    input.Endpoint,
		ActionInput: input.ActionInput,
	})

	return (*RunOutput)(handlerOutput), err
}

type HandlerInput struct {
	Endpoint string
	*ActionInput
}

type HandlerOutput struct {
	*RequestResponse
}

type Request struct {
	Header *RequestHeader `json:"Header,omitempty"`
	Body   *RequestBody   `json:"Body,omitempty"`
}

type RequestHeader struct {
	Action string `json:"Action"`
}

type RequestBody struct {
	*ActionInput
}

type RequestResponse struct {
	Header *RequestResponseHeader
	Body   *RequestResponseBody
}

type RequestResponseHeader struct {
	SessionId string `json:"SessionId,omitempty"`
	Action    string `json:"Action,omitempty"`
	Status    string `json:"Status,omitempty"`
	Error     bool   `json:"Error,omitempty"`
}

type RequestResponseBody struct {
	NomeImob       *string `json:"NomeImob,omitempty"`       // Nome da administradora.
	ImobId         *string `json:"ImobId,omitempty"`         // Identificação da administradora.
	UsuarioId      *string `json:"UsuarioId,omitempty"`      // Identificação do usuário.
	Nome           *string `json:"Nome,omitempty"`           // Nome do usuário.
	Versao         *string `json:"Versao,omitempty"`         // Versão do sistema no servidor.
	ClientIP       *string `json:"Client_IP,omitempty"`      // Endereço IP da estação.
	CodFilial      *int    `json:"CodFilial,omitempty"`      // Código da filial.
	NomeFilial     *string `json:"NomeFilial,omitempty"`     // Nome da filial.
	Cidade         *string `json:"Cidade,omitempty"`         // Cidade da filial.
	Uf             *string `json:"UF,omitempty"`             // UF da filial.
	MaxSessions    *int    `json:"MaxSessions,omitempty"`    // Limite de sessões simultâneas deste usuário.
	ServerDateTime *string `json:"ServerDateTime,omitempty"` // Horário do login no servidor.
}

func handler(input *HandlerInput) (*HandlerOutput, error) {
	request := Request{
		Header: &RequestHeader{
			Action: ACTION,
		},
		Body: &RequestBody{
			input.ActionInput,
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
