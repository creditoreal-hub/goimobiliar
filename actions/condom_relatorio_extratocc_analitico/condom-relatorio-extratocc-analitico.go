package condom_relatorio_extratocc_analitico

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

var ACTION = "CONDOM_RELATORIO_EXTRATOCC_ANALITICO"

type ActionInput struct {
	CodCondominio  *int    `json:"CodCondominio,omitempty"`  // *Código do condomínio.
	Competencia    *string `json:"Competencia,omitempty"`    // *Competência referência da Pasta.
	ResponseFormat *string `json:"Responseformat,omitempty"` // Formato desejado da resposta.
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
	CodCondominio     *int                              `json:"CodCondominio,omitempty"`     // Código do condomínio.
	Competencia       *string                           `json:"Competencia,omitempty"`       // Competência do extrato a gerar.
	Contas            *[]RequestResponseBodyConta       `json:"Contas,omitempty"`            // Informações de cada bloco/conta.
	ResumoSaldos      *[]RequestResponseBodyResumoSaldo `json:"ResumoSaldos,omitempty"`      //
	SaldoGeral        *float64                          `json:"SaldoGeral,omitempty"`        // Saldo geral do condomínio.
	DataProcessamento *string                           `json:"DataProcessamento,omitempty"` // Data e hora do processamento das informações.
}

type RequestResponseBodyConta struct {
	CodBloco           *string                                     `json:"CodBloco,omitempty"`  // Código de bloco/conta.
	NomeBloco          *string                                     `json:"NomeBloco,omitempty"` // Nome de bloco/conta.
	LancamentosCC      *[]RequestResponseBodyContaLancamentoCC     // Lançamentos em conta corrente.
	LancamentosFuturos *[]RequestResponseBodyContaLancamentoFuturo // Lançamentos com vencimentos futuros.
	Resumos            *[]RequestResponseBodyContaResumo           //
	ResumoConta        *RequestResponseBodyContaResumoConta        // Resumo sintético dos lançamentos deste bloco/conta.
	ControleBoletos    *[]RequestResponseBodyContaControleBoletos  //
}

type RequestResponseBodyContaLancamentoCC struct {
	Data         *string  `json:"Data,omitempty"`         // Data do lançamento.
	Historico    *string  `json:"Historico,omitempty"`    // Histórico do lançamento.
	ValorDebito  *float64 `json:"ValorDebito,omitempty"`  // Valor de débito do lançamento.
	ValorCredito *float64 `json:"ValorCredito,omitempty"` // Valor de crébito do lançamento.
	Saldo        *float64 `json:"Saldo,omitempty"`        // Saldo resultante do lançamento.
	NumeroLancto *int     `json:"NumeroLancto,omitempty"` // Número do lançamento.
	CodTaxa      *int     `json:"CodTaxa,omitempty"`      // Código da taxa deste lançamento.
}

type RequestResponseBodyContaLancamentoFuturo struct {
	Data         *string  `json:"Data,omitempty"`         // Data do lançamento.
	Historico    *string  `json:"Historico,omitempty"`    // Histórico do lançamento.
	ValorDebito  *float64 `json:"ValorDebito,omitempty"`  // Valor de débito do lançamento.
	ValorCredito *float64 `json:"ValorCredito,omitempty"` // Valor de crébito do lançamento.
	Saldo        *float64 `json:"Saldo,omitempty"`        // Saldo resultante do lançamento.
	NumeroLancto *int     `json:"NumeroLancto,omitempty"` // Número do lançamento.
	CodTaxa      *int     `json:"CodTaxa,omitempty"`      // Código da taxa deste lançamento.
}

type RequestResponseBodyContaResumo struct {
	Titulo            *string                                           `json:"Titulo,omitempty"`            // Título do resumo de lançamentos.
	LancamentosResumo *[]RequestResponseBodyContaResumoLancamentoResumo `json:"LancamentosResumo,omitempty"` // Lançamentos de resumo.
	SubTotal          *float64                                          `json:"SubTotal,omitempty"`          //Subtotal dos lançamentos.
}

type RequestResponseBodyContaResumoLancamentoResumo struct {
	Historico    *string  `json:"Historico,omitempty"`    // Histórico do lançamento.
	ValorDebito  *float64 `json:"ValorDebito,omitempty"`  // Valor de débito do lançamento.
	ValorCredito *float64 `json:"ValorCredito,omitempty"` // Valor de crébito do lançamento.
}

type RequestResponseBodyContaResumoConta struct {
	SaldoAnterior *float64 `json:"SaldoAnterior,omitempty"` // Saldo de bloco/conta anterior aos lançamentos.
	Despesa       *float64 `json:"Despesa,omitempty"`       // Valor total das despesas.
	Receita       *float64 `json:"Receita,omitempty"`       // Valor total das receitas.
	SaldoFinal    *float64 `json:"SaldoFinal,omitempty"`    // Saldo final de bloco/conta após os lançamentos.
}

type RequestResponseBodyContaControleBoletos struct {
	QtdeBoletos  *int     `json:"QtdeBoletos,omitempty"`  // Quantidade de boletos.
	ValorBoletos *float64 `json:"ValorBoletos,omitempty"` // Valor dos boletos.
	Percentual   *float64 `json:"Percentual,omitempty"`   // Percentual dos boletos em relação ao total.
	Controle     *string  `json:"Controle,omitempty"`     // Identificação do controle.
}

type RequestResponseBodyResumoSaldo struct {
	CodBloco   *string  `json:"CodBloco,omitempty"`   // Código de bloco/conta.
	NomeBloco  *string  `json:"NomeBloco,omitempty"`  // Nome de bloco/conta.
	SaldoBloco *float64 `json:"SaldoBloco,omitempty"` // Saldo resultante dos lançamentos.
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
