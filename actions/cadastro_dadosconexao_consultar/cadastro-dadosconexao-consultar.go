package cadastro_dadosconexao_consultar

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/creditoreal-hub/goimobiliar/consts"
	"github.com/creditoreal-hub/goimobiliar/erros"
	"github.com/creditoreal-hub/goimobiliar/session"
)

var ACTION = "CADASTRO_DADOSCONEXAO_CONSULTAR"

type ActionInput struct {
	Origem                   *string `json:"Origem,omitempty"`                   // *Origem dos Dados de Conexão.
	CodigoOrigem             *int    `json:"CodigoOrigem,omitempty"`             // Código do cadastro de origem vinculado aos Dados de Conexão.
	CodigoOrigemComplementar *string `json:"CodigoOrigemComplementar,omitempty"` // Código complementar do cadastro de origem vinculado aos Dados de Conexão.
	RoboID                   *string `json:"RoboID,omitempty"`                   // Identificação do Robô.
}

type RunMultiInput consts.RunMultiInput[*ActionInput]
type RunMultiOutput consts.RunMultiOutput[*RunOutput]

func RunMulti(input *RunMultiInput) (*RunMultiOutput, error) {
	output := make(RunMultiOutput, 0, len(input.Entries))

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, entry := range input.Entries {
		if input.Parallel {
			wg.Add(1)
			go func(entry *consts.RunMultiInputEntry[*ActionInput]) {
				defer wg.Done()
				mu.Lock()
				output = append(output, runMultiHandler(entry))
				mu.Unlock()
			}(entry)
		} else {
			output = append(output, runMultiHandler(entry))
		}
	}

	if input.Parallel {
		wg.Wait()
	}

	return &output, nil
}

func runMultiHandler(input *consts.RunMultiInputEntry[*ActionInput]) *consts.RunMultiOutputEntry[*RunOutput] {
	outputEntry := consts.RunMultiOutputEntry[*RunOutput]{
		ImobId: input.ImobId,
	}

	sess, err := session.NewSession(&session.NewInput{
		Endpoint: input.Endpoint,
		ImobId:   input.ImobId,
		UserId:   input.UserId,
		UserPass: input.UserPass,
	})
	if err != nil {
		msg := err.Error()

		outputEntry.Success = false
		outputEntry.Error.Message = msg

		return &outputEntry
	}

	defer sess.EndSession()

	handlerOutput, err := Run(&RunInput{
		Session:     sess,
		ActionInput: input.Input,
	})
	if err != nil {
		msg := err.Error()

		outputEntry.Success = false
		outputEntry.Error.Message = msg

		return &outputEntry
	}

	outputEntry.Success = true
	outputEntry.Data = handlerOutput

	return &outputEntry
}

type RunInput HandlerInput
type RunOutput RequestResponseBody

func Run(input *RunInput) (*RunOutput, error) {
	if input.Session == nil {
		return nil, erros.ErrBaseInvalida
	}

	handlerOutput, err := handler(&HandlerInput{
		Session:     input.Session,
		ActionInput: input.ActionInput,
	})
	if err != nil {
		return nil, err
	}

	return (*RunOutput)(handlerOutput.Body), nil
}

type HandlerInput struct {
	Session *session.Session
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
	SessionId string `json:"SessionId,omitempty"`
	Action    string `json:"Action,omitempty"`
}

type RequestBody struct {
	*ActionInput
}

type RequestResponse struct {
	Header *RequestResponseHeader `json:"Header,omitempty"`
	Body   *RequestResponseBody   `json:"Body,omitempty"`
}

type RequestResponseHeader struct {
	SessionId string `json:"SessionId,omitempty"`
	Action    string `json:"Action,omitempty"`
	Status    string `json:"Status,omitempty"`
	Error     bool   `json:"Error,omitempty"`
}

type RequestResponseBody struct {
	DadosConexoes *[]RequestResponseDadosConexao `json:"DadosConexoes,omitempty"` //
}

type RequestResponseDadosConexao struct {
	Origem                   *string `json:"Origem,omitempty"`                   // Origem dos Dados de Conexão.
	CodigoOrigem             *int    `json:"CodigoOrigem,omitempty"`             // Código do cadastro de origem vinculado aos Dados de Conexão.
	CodigoOrigemComplementar *string `json:"CodigoOrigemComplementar,omitempty"` // Código complementar do cadastro de origem vinculado aos Dados de Conexão.
	RoboID                   *string `json:"RoboID,omitempty"`                   // Identificação do Robô.
	RoboNome                 *string `json:"RoboNome,omitempty"`                 // Nome do Robô.
	CodigoFornecedor         *int    `json:"CodigoFornecedor,omitempty"`         // Código do fornecedor.
	NomeFornecedor           *string `json:"NomeFornecedor,omitempty"`           // Nome/Razão Social do fornecedor.
	Login                    *string `json:"Login,omitempty"`                    // Login de acesso ao WebService.
	Senha                    *string `json:"Senha,omitempty"`                    // Senha de acesso ao WebService.
	WebServiceAtivo          *string `json:"WebServiceAtivo,omitempty"`          // Indica se possui WebService ativo.
	WebServiceURL            *string `json:"WebServiceURL,omitempty"`            // Endereço do WebService (URL base sem parâmetros).
	WebServiceComplemento    *string `json:"WebServiceComplemento,omitempty"`    // Complementos da URL base do WebService.
}

func handler(input *HandlerInput) (*HandlerOutput, error) {
	request := Request{
		Header: &RequestHeader{
			SessionId: input.Session.SessionId,
			Action:    ACTION,
		},
		Body: &RequestBody{
			input.ActionInput,
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", input.Session.Endpoint, bytes.NewBuffer(data))
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
