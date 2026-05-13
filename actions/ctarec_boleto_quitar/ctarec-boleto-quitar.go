package ctarec_boleto_quitar

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

var ACTION = "CTAREC_BOLETO_QUITAR"

type ActionInput struct {
	OrigemCobranca            *string  `json:"OrigemCobranca,omitempty"`            // *Locação/Condominio.
	NossoNumero               *string  `json:"NossoNumero,omitempty"`               // *Número de identificação bancário.
	DocCapaId                 *int     `json:"DocCapaId,omitempty"`                 // *Código do boleto no sistema.
	DataPagamento             *string  `json:"DataPagamento,omitempty"`             // *Data do pagamento.
	OrigemQuitacao            *string  `json:"OrigemQuitacao,omitempty"`            // *Origem.
	VlrJuros                  *types.Float64 `json:"VlrJuros,omitempty"`                  // Valor dos juros.
	VlrMulta                  *types.Float64 `json:"VlrMulta,omitempty"`                  // Valor da multa.
	VlrMultaAdministrativa    *types.Float64 `json:"VlrMultaAdministrativa,omitempty"`    // *Valor multa administrativa.
	VlrDescontoAdministrativo *types.Float64 `json:"VlrDescontoAdministrativo,omitempty"` // *Valor desconto administrativo.
	VlrAcrescimoOutros        *types.Float64 `json:"VlrAcrescimoOutros,omitempty"`        // *Valor de acrescimos extras (não incluir multa e ou juros).
	VlrDescontoProprietario   *types.Float64 `json:"VlrDescontoProprietario,omitempty"`   // *Valor do desconto concedido pelo proprietario.
	VlrDescontos              *types.Float64 `json:"VlrDescontos,omitempty"`              // *Valor total dos descontos.
	SeguroConteudo            *string  `json:"SeguroConteudo,omitempty"`            // Pagou o seguro conteúdo. Valor default é 'N'.
	VlrAcrescimos             *types.Float64 `json:"VlrAcrescimos,omitempty"`             // *Valor total dos acrescimos (multa e juros) mais o valor seguro conteúdo.
	VlrPagamento              *types.Float64 `json:"VlrPagamento,omitempty"`              // *Valor pago.
	CodBanco                  *int     `json:"CodBanco,omitempty"`                  // *Código do banco.
	Complemento               *string  `json:"Complemento,omitempty"`               // *
	IdAdmCCDeposito           *int     `json:"IdAdmCCDeposito,omitempty"`           // Id interno da conta corrente que recebeu o deposito.
	DataDeposito              *string  `json:"DataDeposito,omitempty"`              // Data do deposito no banco.
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
	GarantiuCond      *string                          `json:"GarantiuCond,omitempty"`        //
	NaoMultiplicaCoef *string                          `json:"Nao_Multiplica_Coef,omitempty"` //
	Lancamentos       *[]RequestResponseBodyLancamento `json:"Lancamentos,omitempty"`         //
}

type RequestResponseBodyLancamento struct {
	NumeroLancto *int `json:"NumeroLancto,omitempty"` // Número do lançamento.
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
