package cadastro_pessoa_consultar_vinculo

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/itispx/goimobiliar/consts"
	"github.com/itispx/goimobiliar/erros"
	"github.com/itispx/goimobiliar/session"
)

var ACTION = "CADASTRO_PESSOA_CONSULTAR_VINCULO"

type ActionInput struct {
	CodPessoa   *int    `json:"CodPessoa,omitempty"`   // *Código da pessoa.
	TipoVinculo *string `json:"TipoVinculo,omitempty"` // Tipo do vínculo da pessoa. Valor default é 'TODOS'.
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
	CodPessoa *string                    `json:"CodPessoa,omitempty"` // Código da pessoa.
	Tipos     *[]RequestResponseBodyTipo `json:"Tipos,omitempty"`     //
}

type RequestResponseBodyTipo struct {
	TipoVinculo *string                           `json:"TipoVinculo,omitempty"` // Tipo do vínculo da pessoa.
	NomeVinculo *string                           `json:"NomeVinculo,omitempty"` // Nome do vínculo da pessoa.
	Vinculos    *[]RequestResponseBodyTipoVinculo `json:"Vinculos,omitempty"`    //
}

type RequestResponseBodyTipoVinculo struct {
	CodCondominio *int    `json:"CodCondominio,omitempty"` // Código do condomínio (quando a busca for Condômino/Síndico).
	CodBloco      *string `json:"CodBloco,omitempty"`      // Código do bloco do condomínio (quando a busca for Condômino/Síndico).
	CodEconomia   *string `json:"CodEconomia,omitempty"`   // Código da economia/unidade no bloco (quando a busca for Condômino).
	IdEconomia    *int    `json:"IdEconomia,omitempty"`    // Identificação da economia no sistema (quando a busca for Condômino).
	CodImovel     *int    `json:"CodImovel,omitempty"`     // Código do imóvel(quando a busca for Proprietário/Locatário/Procurador/Beneficiario/Fiador).
	CodContrato   *int    `json:"CodContrato,omitempty"`   // Código do contrato de locação deste imóvel (quando a busca for Locatário).
	Descricao     *string `json:"Descricao,omitempty"`     // Descrição do vínculo.
	Situacao      *string `json:"Situacao,omitempty"`      // Situacao do vínculo.
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
