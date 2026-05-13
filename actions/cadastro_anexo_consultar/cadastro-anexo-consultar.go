package cadastro_anexo_consultar

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

var ACTION = "CADASTRO_ANEXO_CONSULTAR"

type ActionInput struct {
	CodAnexo *int `json:"CodAnexo,omitempty"` // *Código do anexo.
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
	CodAnexo      *int                          `json:"CodAnexo,omitempty"`      // Código do anexo.
	TipoAnexo     *int                          `json:"TipoAnexo,omitempty"`     // Código do cadastro de anexo que indica o tipo dos arquivos.
	TipoOrigem    *string                       `json:"TipoOrigem,omitempty"`    // Código do cadastro de origem vinculado ao anexo.
	CodOrigem     *int                          `json:"CodOrigem,omitempty"`     // Código do cadastro de origem vinculado ao anexo.
	SubCodOrigem  *string                       `json:"SubCodOrigem,omitempty"`  // Subcódigo do cadastro de origem vinculado ao anexo.
	Descricao     *string                       `json:"Descricao,omitempty"`     // Descrição do Anexo.
	DataAlteracao *string                       `json:"DataAlteracao,omitempty"` // Data da última alteração do anexo.
	EnviaSite     *string                       `json:"EnviaSite,omitempty"`     // Habilitado para enviar para o site.
	DataEnviaSite *string                       `json:"DataEnviaSite,omitempty"` // Data prevista para enviar para o site.
	TotalArquivos *int                          `json:"TotalArquivos,omitempty"` // Total de arquivos encontrados na consulta.
	Arquivos      *[]RequestResponseBodyArquivo `json:"Arquivos,omitempty"`      //
}

type RequestResponseBodyArquivo struct {
	ArquivoNome     *string `json:"ArquivoNome,omitempty"`     // Nome do arquivo.
	ArquivoTamanho  *string `json:"ArquivoTamanho,omitempty"`  // Tamanho do arquivo em bytes.
	ArquivoDataHora *string `json:"ArquivoDataHora,omitempty"` // Data e hora da última modificação do arquivo no formato: AAAA-MM-DD-hh-mm-ss.
	URL             *string `json:"URL,omitempty"`             // URL para download do arquivo.
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
