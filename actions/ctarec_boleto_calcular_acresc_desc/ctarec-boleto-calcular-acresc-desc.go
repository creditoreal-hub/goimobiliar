package ctarec_boleto_calcular_acresc_desc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/creditoreal-hub/goimobiliar/consts"
	"github.com/creditoreal-hub/goimobiliar/erros"
	"github.com/creditoreal-hub/goimobiliar/session"
	"github.com/creditoreal-hub/goimobiliar/types"
)

var ACTION = "CTAREC_BOLETO_CALCULAR_ACRESC_DESC"

type ActionInput struct {
	NossoNumero     *string `json:"NossoNumero,omitempty"`     // Número de identificação bancário.
	DocCapaId       *int    `json:"DocCapaId,omitempty"`       // Código do boleto no sistema.
	DataLimitePagto *string `json:"DataLimitePagto,omitempty"` // Data limite de pagamento.
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
	VlrDocumento     *types.Float64 `json:"VlrDocumento,omitempty"`     // Valor do documento sem o seguro conteúdo, multa e juros.
	VlrDocumentoOrig *types.Float64 `json:"VlrDocumentoOrig,omitempty"` // Valor original do documento.
	Juros            *types.Float64 `json:"Juros,omitempty"`            // Valor de juros do documento.
	Multa            *types.Float64 `json:"Multa,omitempty"`            // Valor de multa do documento.
	VlrCorrecao      *types.Float64 `json:"VlrCorrecao,omitempty"`      // Valor de correçao do documento.
	MultaProp        *types.Float64 `json:"MultaProp,omitempty"`        // Valor de multa do proprietário.
	VlrCorrecaoProp  *types.Float64 `json:"VlrCorrecaoProp,omitempty"`  // Valor de correçao do proprietário.
	DescontoProp     *types.Float64 `json:"DescontoProp,omitempty"`     // Valor de desconto do proprietário.
	DescontoAdm      *types.Float64 `json:"DescontoAdm,omitempty"`      // Valor de desconto da administradora.
	VlrCorrigido     *types.Float64 `json:"VlrCorrigido,omitempty"`     // Valor corrigido do documento.
	NossoNumero      *string  `json:"NossoNumero,omitempty"`      // Número de identificação bancário.
	DocCapaId        *int     `json:"DocCapaId,omitempty"`        // Código do boleto no sistema.
	OrigemCobranca   *string  `json:"OrigemCobranca,omitempty"`   // Locação/Condominio.
	VlrSegCont       *types.Float64 `json:"VlrSegCont,omitempty"`       // Valor do seguro conteúdo.
	ErroIndice       *string  `json:"ErroIndice,omitempty"`       // Mensagem de erro.
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
