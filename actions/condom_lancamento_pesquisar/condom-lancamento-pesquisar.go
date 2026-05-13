package condom_lancamento_pesquisar

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

var ACTION = "CONDOM_LANCAMENTO_INCLUIR"

type ActionInput struct {
	CodCondominio       *int                   `json:"CodCondominio,omitempty"`       // *Código do condomínio.
	CodBloco            *string                `json:"CodBloco,omitempty"`            // Código do bloco do condomínio.
	CodBlocoBase        *string                `json:"CodBlocoBase,omitempty"`        // Bloco base/principal do condomínio.
	Competencia         *string                `json:"Competencia,omitempty"`         // *Competência para a qual o lançamento será lançado.
	Valor               *types.Float64               `json:"Valor,omitempty"`               // *Valor do lançamento.
	Complemento         *string                `json:"Complemento,omitempty"`         // *Complemento descritivo do lançamento.
	CodTaxa             *int                   `json:"CodTaxa,omitempty"`             // *Código da taxa que classifica este lançamento.
	Origem              *string                `json:"Origem,omitempty"`              // Origem do lançamento. Valor default é 'M'.
	CompetenciaReajuste *string                `json:"CompetenciaReajuste,omitempty"` // Competência do reajuste do lançamento.
	PercentualReajuste  *types.Float64               `json:"PercentualReajuste,omitempty"`  // Percentual de reajuste do lançamento. Valor default é '0'.
	DebitoCredito       *string                `json:"DebitoCredito,omitempty"`       // Indica se o lançamento é de crédito ou de débito. Valor default é 'D'.
	TipoLancamento      *string                `json:"TipoLancamento,omitempty"`      // Tipo de lançamento. Valor default é 'I'.
	NumeroParcela       *int                   `json:"NumeroParcela,omitempty"`       // *Número da parcela.
	TotalParcelas       *int                   `json:"TotalParcelas,omitempty"`       // *Número total de parcelas.
	TipoDocumento       *string                `json:"TipoDocumento,omitempty"`       // Tipo de boleto/DOC. Valor default é 'N'.
	DataVencimentoExtra *string                `json:"DataVencimentoExtra,omitempty"` // Data de vencimento se tipo do documento for extra (TipoDocumento='E').
	DocAtrasado         *string                `json:"DocAtrasado,omitempty"`         // Indica se o DOC/boleto é atrasado. Valor default é 'N'.
	DebitarLocatario    *string                `json:"DebitarLocatario,omitempty"`    // Indica se é para debitar o locatário. Valor default é 'N'.
	Economias           *[]ActionInputEconomia `json:"Economias,omitempty"`           // Lista de economias a lançar quando o tipo de lançamento for individual (TipoLancamento='I').
}

type ActionInputEconomia struct {
	IdEconomia *int `json:"IdEconomia,omitempty"` // *Chave principal da economia/unidade.
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
	LanctoCondId *int `json:"LanctoCondId,omitempty"` // Código do lançamento de condomínio.
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
