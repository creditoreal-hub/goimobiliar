package ctapag_relatorio_conferencia

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

var ACTION = "CTAPAG_RELATORIO_CONFERENCIA"

type ActionInput struct {
	Rotina                *string `json:"Rotina,omitempty"`                // *Seleciona qual rotina de lançamentos relacionar.
	Competencia           *string `json:"Competencia,omitempty"`           // *Competência do relatório de conferência.
	CodFilial             *int    `json:"CodFilial,omitempty"`             // Código da filial. Valor default é '000'.
	DataVencimentoInicial *string `json:"DataVencimentoInicial,omitempty"` // Data de vencimento inicial.
	DataVencimentoFinal   *string `json:"DataVencimentoFinal,omitempty"`   // Data de vencimento final.
	CodFornecedor         *int    `json:"CodFornecedor,omitempty"`         // Código de fornecedor desta filial.
	CodTaxa               *int    `json:"CodTaxa,omitempty"`               // Código da taxa que classifica este lançamento.
	PagtoBanco            *string `json:"PagtoBanco,omitempty"`            // Listar lançamentos com forma de pagamento por banco (cheque). Valor default é 'N'.
	PagtoCaixa            *string `json:"PagtoCaixa,omitempty"`            // Listar lançamentos com forma de pagamento por ciaxa (Cheque). Valor default é 'N'.
	PagtoDinheiro         *string `json:"PagtoDinheiro,omitempty"`         // Listar lançamentos com forma de pagamento por dinheiro. Valor default é 'N'.
	PagtoOrdem            *string `json:"PagtoOrdem,omitempty"`            // Listar lançamentos com forma de pagamento por ordem de pagamento. Valor default é 'N'.
	PagtoLiqTitulos       *string `json:"PagtoLiqTitulos,omitempty"`       // Listar lançamentos com forma de pagamento por liquidação de títulos. Valor default é 'N'.
	PagtoLiqTitulosAgrup  *string `json:"PagtoLiqTitulosAgrup,omitempty"`  // Listar lançamentos com forma de pagamento por liquidação de títulos agrupados. Valor default é 'N'.
	PagtoPIXTransf        *string `json:"PagtoPIXTransf,omitempty"`        // Listar lançamentos com forma de pagamento por transferência de PIX. Valor default é 'N'.
	PagtoPIXQrCode        *string `json:"PagtoPIXQrCode,omitempty"`        // Listar lançamentos com forma de pagamento por QRcode de PIX. Valor default é 'N'.
	PagtoCredConta        *string `json:"PagtoCredConta,omitempty"`        // Listar lançamentos com forma de pagamento por crédito em conta. Valor default é 'N'.
	PagtoDecAutomatico    *string `json:"PagtoDecAutomatico,omitempty"`    // Listar lançamentos com forma de pagamento por débito automático. Valor default é 'N'.
	NaoPagar              *string `json:"NaoPagar,omitempty"`              // Listar somente lançamentos marcados para não pagar. Valor default é 'N'.
	SoQuitados            *string `json:"SoQuitados,omitempty"`            // Listar somente lançamentos quitados. Valor default é 'N'.
	SoComDiferenca        *string `json:"SoComDiferenca,omitempty"`        // Listar somente registros com diferença.
	ResponseFormat        *string `json:"ResponseFormat,omitempty"`        // Formato desejado da resposta.
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
	Competencia   *string                           `json:"Competencia,omitempty"`   // Competência do relatório de conferência.
	Lancamentos   *[]RequestResponseBodyLancamento  `json:"Lancamentos,omitempty"`   //
	TotaisLanctos *RequestResponseBodyTotaisLanctos `json:"TotaisLanctos,omitempty"` //
	Resumos       *[]RequestResponseBodyResumo      `json:"Resumos,omitempty"`       //
}

type RequestResponseBodyLancamento struct {
	NumeroLancto   *int     `json:"NumeroLancto,omitempty"`   // Número do lançamento.
	CodImovel      *int     `json:"CodImovel,omitempty"`      // Código do imóvel.
	Descricao      *string  `json:"Descricao,omitempty"`      // Descrição do lançamento/item.
	DataVencimento *string  `json:"DataVencimento,omitempty"` // Data de vencimento do lançamento.
	DataPagamento  *string  `json:"DataPagamento,omitempty"`  // Data de pagamento do lançamento (quando quitado).
	OrigemCobranca *string  `json:"OrigemCobranca,omitempty"` // Origem da cobrança do lançamento.
	Valor          *float64 `json:"Valor,omitempty"`          // Valor do(s) lançamento(s).
	ValorCobrado   *float64 `json:"ValorCobrado,omitempty"`   // Valor cobrado.
	Diferenca      *float64 `json:"Diferenca,omitempty"`      // Diferença na cobrança do(s) lançamento(s).
}

type RequestResponseBodyTotaisLanctos struct {
	Valor        *float64 `json:"Valor,omitempty"`        // Valor do(s) lançamento(s).
	ValorCobrado *float64 `json:"ValorCobrado,omitempty"` // Valor cobrado.
	Diferenca    *float64 `json:"Diferenca,omitempty"`    // Diferença na cobrança do(s) lançamento(s).
}

type RequestResponseBodyResumo struct {
	Titulo       *string                                `json:"Titulo,omitempty"`       // Título do resumo.
	Itens        *[]RequestResponseBodyResumoItens      `json:"Itens,omitempty"`        //
	TotaisResumo *RequestResponseBodyResumoTotaisResumo `json:"TotaisResumo,omitempty"` //
}

type RequestResponseBodyResumoItens struct {
	Descricao    *string  `json:"Descricao,omitempty"`    // Descrição do lançamento/item.
	Quantidade   *int     `json:"Quantidade,omitempty"`   // Quantidade de lançamentos/itens.
	Valor        *float64 `json:"Valor,omitempty"`        // Valor do(s) lançamento(s).
	ValorCobrado *float64 `json:"ValorCobrado,omitempty"` // Valor cobrado.
	Diferenca    *float64 `json:"Diferenca,omitempty"`    // Diferença na cobrança do(s) lançamento(s).
}

type RequestResponseBodyResumoTotaisResumo struct {
	Quantidade   *int     `json:"Quantidade,omitempty"`   // Quantidade de lançamentos/itens.
	Valor        *float64 `json:"Valor,omitempty"`        // Valor do(s) lançamento(s).
	ValorCobrado *float64 `json:"ValorCobrado,omitempty"` // Valor cobrado.
	Diferenca    *float64 `json:"Diferenca,omitempty"`    // Diferença na cobrança do(s) lançamento(s).
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
