package locacao_lancto_cond_incluir

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

var ACTION = "LOCACAO_LANCTO_COND_INCLUIR"

type ActionInput struct {
	TestaAditivoLocacao         *string  `json:"TestaAditivoLocacao,omitempty"`         // *Testar existência de registro válido na tabela 'aditivo de locação'.
	DataPagamento               *string  `json:"DataPagamento,omitempty"`               // Data de pagamento do lançamento (quando quitado).
	LancarContaCorrenteLocacao  *string  `json:"LancarContaCorrenteLocacao,omitempty"`  // *Lançar na conta corrente de locação.
	CompetenciaLocacao          *string  `json:"CompetenciaLocacao,omitempty"`          // *Competência.
	TipoBoleto                  *string  `json:"TipoBoleto,omitempty"`                  // *Tipo de boleto para lançar.
	CodImovel                   *int     `json:"CodImovel,omitempty"`                   // *Código do imóvel.
	CodContratoLoc              *int     `json:"CodContratoLoc,omitempty"`              // *Código do contrato de locação deste imóvel.
	Competencia                 *string  `json:"Competencia,omitempty"`                 // *Competência.
	Complemento                 *string  `json:"Complemento,omitempty"`                 // Complemento.
	CodTaxa                     *int     `json:"CodTaxa,omitempty"`                     // *Código da taxa que classifica este lançamento.
	CobrarLocatarioProprietario *string  `json:"CobrarLocatarioProprietario,omitempty"` // *Cobrar do Locatario ou proprietario.
	TotalParcelas               *int     `json:"TotalParcelas,omitempty"`               // *Quantidade total de parcelas.
	NumeroParcela               *int     `json:"NumeroParcela,omitempty"`               // *Número da parcela do lançamento.
	ValorReal                   *types.Float64 `json:"ValorReal,omitempty"`                   // *Valor real.
	DataVencimento              *string  `json:"DataVencimento,omitempty"`              // *Data de vencimento do lançamento.
	DataVigInicial              *string  `json:"DataVigInicial,omitempty"`              // Data inicial da vigência do contrato.
	CodBarras                   *string  `json:"CodBarras,omitempty"`                   // Código de barras do boleto.
	CodPessoaPagador            *int     `json:"CodPessoaPagador,omitempty"`            // Código do pagador no cadastro de pessoas.
	NomePagador                 *string  `json:"NomePagador,omitempty"`                 // Nome do beneficiário. (Para liquidação de títulos se este for diferente do condomínio).
	CpfCnpjPagador              *int     `json:"CpfCnpjPagador,omitempty"`              // CPF ou CNPJ do pagador. (Para liquidação de títulos se este for diferente do condomínio).
	TipoPessoaPagador           *string  `json:"TipoPessoaPagador,omitempty"`           // Tipo de pessoa do pagador. (Para liquidação de títulos se este for diferente do condomínio).
	CodPessoaBenef              *int     `json:"CodPessoaBenef,omitempty"`              // Código de pessoa do beneficiário.
	NomeBeneficiario            *string  `json:"NomeBeneficiario,omitempty"`            // Nome do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	CpfCnpjBeneficiario         *int     `json:"CpfCnpjBeneficiario,omitempty"`         // CPF ou CNPJ do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	TipoPessoaBeneficiario      *string  `json:"TipoPessoaBeneficiario,omitempty"`      // Tipo de pessoa do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	NoDemonstrativo             *string  `json:"NoDemonstrativo,omitempty"`             // Indica a forma de lançamento no demonstrativo.
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
	NumeroLancto     *int `json:"NumeroLancto,omitempty"`     // Número do lançamento.
	NumeroLanctoItem *int `json:"NumeroLanctoItem,omitempty"` // Número do lançamento.
	CodContratoLoc   *int `json:"CodContratoLoc,omitempty"`   // Código do contrato de locação deste imóvel.
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
