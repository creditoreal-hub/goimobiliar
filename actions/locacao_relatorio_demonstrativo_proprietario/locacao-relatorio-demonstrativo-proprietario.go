package locacao_relatorio_demonstrativo_proprietario

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

var ACTION = "LOCACAO_RELATORIO_DEMONSTRATIVO_PROPRIETARIO"

type ActionInput struct {
	CodPessoa      *int    `json:"CodPessoa,omitempty"`      // *Código de pessoa do proprietário.
	Competencia    *string `json:"Competencia,omitempty"`    // *Competência do demonstrativo.
	CodFilial      *int    `json:"CodFilial,omitempty"`      // Código da filial. Valor default é '001'.
	ResponseFormat *string `json:"ResponseFormat,omitempty"` // Formato desejado da resposta.
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
	Proprietarios *[]RequestResponseBodyProprietario `json:"Proprietarios,omitempty"` //
}

type RequestResponseBodyProprietario struct {
	CodPessoa                           *int                                                     `json:"CodPessoa,omitempty"`                           // Código de pessoa do proprietário.
	Nome                                *string                                                  `json:"Nome,omitempty"`                                // Nome do proprietário.
	TipoPessoa                          *string                                                  `json:"TipoPessoa,omitempty"`                          // Tipo da pessoa.
	CpfCnpj                             *types.CpfCnpj                                           `json:"CpfCnpj,omitempty"`                             // Se for tipo de pessoa física o valor é um CPF. Se for tipo de pessoa jurídica o valor é um CNPJ. Se o tipo de pessoa não for informado então este campo é vazio.
	Competencia                         *string                                                  `json:"Competencia,omitempty"`                         // Competência do demonstrativo.
	FilialNome                          *string                                                  `json:"FilialNome,omitempty"`                          // Nome da filial.
	Titulo                              *string                                                  `json:"Titulo,omitempty"`                              // Título do relatório.
	FormaPagto                          *string                                                  `json:"FormaPagto,omitempty"`                          // Forma de pagamento ao proprietário.
	CodBanco                            *int                                                     `json:"CodBanco,omitempty"`                            // Código do banco onde esta pessoa tem conta bancária.
	CodAgencia                          *int                                                     `json:"CodAgencia,omitempty"`                          // Código da agencia onde esta pessoa tem conta bancária.
	TipoConta                           *string                                                  `json:"TipoConta,omitempty"`                           // Tipo da conta bancária desta pessoa.
	ContaCorrente                       *string                                                  `json:"ContaCorrente,omitempty"`                       // Número da conta bancária desta pessoa.
	AssessorNome                        *string                                                  `json:"AssessorNome,omitempty"`                        // Nome do assessor/gestor.
	CodAssessor                         *int                                                     `json:"CodAssessor,omitempty"`                         // Código do assessor/gestor.
	AssessorEmail                       *string                                                  `json:"AssessorEmail,omitempty"`                       // Email do assessor/gestor.
	AssessorTelefone                    *string                                                  `json:"AssessorTelefone,omitempty"`                    // Telefone do assessor/gestor.
	CodPessoaBenef                      *int                                                     `json:"CodPessoaBenef,omitempty"`                      // Código de pessoa do beneficiário.
	NomeBenef                           *string                                                  `json:"NomeBenef,omitempty"`                           // Nome do beneficiário.
	LancamentosProprietario             *[]RequestResponseBodyProprietarioLancamentoProprietario `json:"LancamentosProprietario,omitempty"`             // Lançamentos do proprietário.
	ValorLiquidoLancamentosProprietario *types.Float64                                           `json:"ValorLiquidoLancamentosProprietario,omitempty"` // Valor liquido dos lançamentos do proprietário.
	Imoveis                             *[]RequestResponseBodyProprietarioLancamentoProprietario `json:"Imoveis,omitempty"`                             // Informações de cada imóvel.
	ResumoTaxas                         *[]RequestResponseBodyProprietarioResumoTaxa             `json:"ResumoTaxas,omitempty"`                         // Resumo de valores por taxa.
	ResumoGeral                         *RequestResponseBodyProprietarioResumoGeral              `json:"ResumoGeral,omitempty"`                         // Resumo geral de valores.
	Pagamentos                          *[]RequestResponseBodyProprietarioPagamento              `json:"Pagamentos,omitempty"`                          // Pagamentos realizados ao proprietário.
}

type RequestResponseBodyProprietarioLancamentoProprietario struct {
	Data         *string        `json:"Data,omitempty"`         // Data do lançamento.
	Competencia  *string        `json:"Competencia,omitempty"`  // Competência do demonstrativo.
	Historico    *string        `json:"Historico,omitempty"`    // Histórico do lançamento.
	ValorDebito  *types.Float64 `json:"ValorDebito,omitempty"`  // Valor de débito do lançamento.
	ValorCredito *types.Float64 `json:"ValorCredito,omitempty"` // Valor de crébito do lançamento.
	NumeroLancto *int           `json:"NumeroLancto,omitempty"` // Número do lançamento.
}

type RequestResponseBodyProprietarioImovel struct {
	CodImovel                     *int                                                     `json:"CodImovel,omitempty"`                     // Código do imóvel.
	Endereco                      *string                                                  `json:"Endereco,omitempty"`                      // Endereço do imóvel.
	Observacao                    *string                                                  `json:"Observacao,omitempty"`                    // Observações referente ao imóvel.
	Locatarios                    *[]RequestResponseBodyProprietarioImovelLocatario        `json:"Locatarios,omitempty"`                    // Locatários do imóvel.
	LancamentosImovel             *[]RequestResponseBodyProprietarioImovelLancamentoImovel `json:"LancamentosImovel,omitempty"`             // Lançamentos do imóvel.
	ValorLiquidoLancamentosImovel *types.Float64                                           `json:"ValorLiquidoLancamentosImovel,omitempty"` // Valor liquido dos lançamentos do imóvel.
}

type RequestResponseBodyProprietarioImovelLocatario struct {
	CodPessoaLocat   *int           `json:"CodPessoaLocat,omitempty"`   // Código de pessoa do locatário principal.
	NomeLocat        *string        `json:"NomeLocat,omitempty"`        // Nome do locatário.
	CpfCnpjLocat     *types.CpfCnpj `json:"CpfCnpjLocat,omitempty"`     // CPF ou CNPJ do locatário.
	TipoPessoaLocat  *string        `json:"TipoPessoaLocat,omitempty"`  // Tipo da pessoa.
	DataDesocupacao  *string        `json:"DataDesocupacao,omitempty"`  // Data de desocupação.
	DataProxReajuste *string        `json:"DataProxReajuste,omitempty"` // Data do próximo reajuste.
	IndiceReajuste   *string        `json:"IndiceReajuste,omitempty"`   // Indice do reajuste.
	DataVigInicial   *string        `json:"DataVigInicial,omitempty"`   // Data inicial da vigência do contrato.
	GaranteAluguel   *string        `json:"GaranteAluguel,omitempty"`   // Indica se tem garantia de aluguel.
	GaranteEncargos  *string        `json:"GaranteEncargos,omitempty"`  // Indica se tem garantia de encargos
}

type RequestResponseBodyProprietarioImovelLancamentoImovel struct {
	Data         *string        `json:"Data,omitempty"`         // Data do lançamento.
	Competencia  *string        `json:"Competencia,omitempty"`  // Competência do demonstrativo.
	Historico    *string        `json:"Historico,omitempty"`    // Histórico do lançamento.
	ValorDebito  *types.Float64 `json:"ValorDebito,omitempty"`  // Valor de débito do lançamento.
	ValorCredito *types.Float64 `json:"ValorCredito,omitempty"` // Valor de crébito do lançamento.
	NumeroLancto *int           `json:"NumeroLancto,omitempty"` // Número do lançamento.
}

type RequestResponseBodyProprietarioResumoTaxa struct {
	Historico    *string        `json:"Historico,omitempty"`    // Histórico do lançamento.
	ValorDebito  *types.Float64 `json:"ValorDebito,omitempty"`  // Valor de débito do lançamento.
	ValorCredito *types.Float64 `json:"ValorCredito,omitempty"` // Valor de crébito do lançamento.
}

type RequestResponseBodyProprietarioResumoGeral struct {
	TotalTaxasCredito    *types.Float64 `json:"TotalTaxasCredito,omitempty"`    // Total das taxas de crédito.
	TotalTaxasDebito     *types.Float64 `json:"TotalTaxasDebito,omitempty"`     // Total das taxas de débito.
	TotalTaxasLiquido    *types.Float64 `json:"TotalTaxasLiquido,omitempty"`    // Total líquido das taxas.
	SaldoAnteriorCredito *types.Float64 `json:"SaldoAnteriorCredito,omitempty"` // Valor de crédito no saldo anterior.
	SaldoAnteriorDebito  *types.Float64 `json:"SaldoAnteriorDebito,omitempty"`  // Valor de débito no saldo anterior.
	SaldoAnteriorLiquido *types.Float64 `json:"SaldoAnteriorLiquido,omitempty"` // Valor líquido do saldo anterior.
	SaldoFinal           *types.Float64 `json:"SaldoFinal,omitempty"`           // Saldo final líquido.
	ValorImpostos        *types.Float64 `json:"ValorImpostos,omitempty"`        // Valor dos impostos.
}

type RequestResponseBodyProprietarioPagamento struct {
	Data     *string        `json:"Data,omitempty"`     // Data do lançamento.
	Valor    *types.Float64 `json:"Valor,omitempty"`    // Valor pago ao proprietário.
	Situacao *string        `json:"Situacao,omitempty"` // Situação do pagamento.
	Saldo    *types.Float64 `json:"Saldo,omitempty"`    // Saldo do proprietário.
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
