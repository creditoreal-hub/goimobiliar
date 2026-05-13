package cadastro_tarefa_alterar

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

var ACTION = "CADASTRO_TAREFA_ALTERAR"

type ActionInput struct {
	CodTarefa     *int    `json:"CodTarefa,omitempty"`     // *Código da tarefa.
	CodCategoria  *int    `json:"CodCategoria,omitempty"`  // Código da categoria da tarefa.
	CodTicket     *int    `json:"CodTicket,omitempty"`     // Código do chamado da integração.
	AlocadaPara   *string `json:"AlocadaPara,omitempty"`   // *ID do usuário que está com a tarefa.
	CodAssunto    *int    `json:"CodAssunto,omitempty"`    // Código do assunto cadastrado no sistema.
	Assunto       *string `json:"Assunto,omitempty"`       // Assunto da tarefa.
	Texto         *string `json:"Texto,omitempty"`         // Texto da tarefa.
	CodContato    *int    `json:"CodContato,omitempty"`    // Código do contato cadastrado no sistema.
	TipoContato   *string `json:"TipoContato,omitempty"`   // Tipo do contato.
	TextoContato  *string `json:"TextoContato,omitempty"`  // Texto do contato.
	DataPrevisao  *string `json:"DataPrevisao,omitempty"`  // Data prevista para a finalização da tarefa.
	DataConclusao *string `json:"DataConclusao,omitempty"` // Data da conclusão da tarefa.
	CodSituacao   *int    `json:"CodSituacao,omitempty"`   // Código da situação da tarefa.
	CodPrioridade *int    `json:"CodPrioridade,omitempty"` // Código da prioridade da tarefa (deve existir no cadastro).
	Percentual    *int    `json:"Percentual,omitempty"`    // Percentual do andamento da tarefa.
	Executor      *string `json:"Executor,omitempty"`      // Texto livre para identificar o responsável pela tarefa.
	Custo         *string `json:"Custo,omitempty"`         // Texto livre para indicar o custo da tarefa.
	CodFornecedor *int    `json:"CodFornecedor,omitempty"` // Código do fornecedor.
	TemLembrete   *string `json:"TemLembrete,omitempty"`   // Indica se a tarefa deve ser lembrada.
	DataLembrete  *string `json:"DataLembrete,omitempty"`  // Data e hora para lembrar a tarefa.
	TextoLembrete *string `json:"TextoLembrete,omitempty"` // Texto livre para lembrar da tarefa
}

type ActionInputAnexo struct {
	DescricaoArquivo string `json:"DescricaoArquivo,omitempty"` // *Descrição do arquivo de anexo que será armazenado no sistema.
	UrlArquivo       string `json:"UrlArquivo,omitempty"`       // Caminho completo (URL) do arquivo para download. Os tipos aceitos são imagens (jpg) e documentos (pdf/zip/doc/eml). Exemplo: https://servidor.com.br/pasta/subpasta/arquivo.pdf.
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
	CodTarefa *int `json:"CodTarefa,omitempty"` // Código da tarefa.
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
