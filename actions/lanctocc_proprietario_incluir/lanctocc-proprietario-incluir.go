package lanctocc_proprietario_incluir

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

var ACTION = "LANCTOCC_PROPRIETARIO_INCLUIR"

type ActionInput struct {
	CodPessoaProprietario *int     `json:"CodPessoaProprietario,omitempty"` // *Código do proprietário.
	DcCcImovel            *string  `json:"DcCcImovel,omitempty"`            // Débito ou crédito na conta corrente do imóvel. Valor default é ' '.
	DcReciboProprietario  *string  `json:"DcReciboProprietario,omitempty"`  // Débito ou crédito no recibo de proprietário. Valor default é ' '.
	NoDemonstrativo       *string  `json:"NoDemonstrativo,omitempty"`       // Indicação da forma de lançamento no demonstrativo. Valor default é 'S'.
	CodFilial             *string  `json:"CodFilial,omitempty"`             // *Código da filial do lançamento.
	Competencia           *string  `json:"Competencia,omitempty"`           // *Competência do lançamento no formato 'YYYYMM'.
	DataPagamento         *string  `json:"DataPagamento,omitempty"`         // *Data de pagamento do lançamento (quando quitado).
	TipoDocumento         *string  `json:"TipoDocumento,omitempty"`         // *Tipo de documento do lançamento.
	NumeroDocumento       *string  `json:"NumeroDocumento,omitempty"`       // *Número do documento do fornecedor.
	Complemento           *string  `json:"Complemento,omitempty"`           // Complemento descritivo do lançamento.
	ComplementoAdicional1 *string  `json:"ComplementoAdicional1,omitempty"` // Informação de complemento extra.
	ComplementoAdicional2 *string  `json:"ComplementoAdicional2,omitempty"` // Informação de complemento extra.
	ComplementoAdicional3 *string  `json:"ComplementoAdicional3,omitempty"` // Informação de complemento extra.
	ComplementoAdicional4 *string  `json:"ComplementoAdicional4,omitempty"` // Informação de complemento extra.
	ComplementoAdicional5 *string  `json:"ComplementoAdicional5,omitempty"` // Informação de complemento extra.
	ComplementoAdicional6 *string  `json:"ComplementoAdicional6,omitempty"` // Informação de complemento extra.
	ComplementoAdicional7 *string  `json:"ComplementoAdicional7,omitempty"` // Informação de complemento extra.
	ComplementoAdicional8 *string  `json:"ComplementoAdicional8,omitempty"` // Informação de complemento extra.
	CodContraPartida      *int     `json:"CodContraPartida,omitempty"`      // Código da conta de contra partida cadastrada no plano de contas da administradora.
	LancaNaViradaParcelas *string  `json:"LancaNaViradaParcelas,omitempty"` // *Lançamento automático na virada de parcelas.
	NumeroParcela         *int     `json:"NumeroParcela,omitempty"`         // Número da parcela do lançamento. Valor default é '1'.
	TotalParcelas         *int     `json:"TotalParcelas,omitempty"`         // Quantidade total de parcelas. Valor default é '1'.
	CodTaxa               *int     `json:"CodTaxa,omitempty"`               // *Código da taxa que classifica este lançamento.
	Valor                 *float64 `json:"Valor,omitempty"`                 // *Valor total do lançamento.
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
