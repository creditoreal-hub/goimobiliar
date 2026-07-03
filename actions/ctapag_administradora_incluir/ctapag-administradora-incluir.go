package ctapag_administradora_incluir

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

var ACTION = "CTAPAG_ADMINISTRADORA_INCLUIR"

type ActionInput struct {
	CodPlanoContaAdm           *int           `json:"CodPlanoContaAdm,omitempty"`           // *Código da conta no plano de contas da administradora (se origem for 'A').
	CodAgencia                 *int           `json:"CodAgencia,omitempty"`                 // Código da agência/loja. Valor default é ''.
	CodCentroCusto             *int           `json:"CodCentroCusto,omitempty"`             // Código do centro de custo da administradora (se origem for 'A'). Valor default é '0'.
	CodFilial                  *string        `json:"CodFilial,omitempty"`                  // *Código da filial do lançamento.
	Competencia                *string        `json:"Competencia,omitempty"`                // Competência do lançamento no formato 'YYYYMM'.
	CodFornecedor              *int           `json:"CodFornecedor,omitempty"`              // Código do fornecedor do lançamento.
	CodPessoaFavorecido        *int           `json:"CodPessoaFavorecido,omitempty"`        // Código do favorecido no cadastro de pessoas.
	NomeFavorecido             *string        `json:"NomeFavorecido,omitempty"`             // Nome do favorecido. Valor default é ' '.
	CodTaxa                    *int           `json:"CodTaxa,omitempty"`                    // *Código da taxa que classifica este lançamento.
	NumeroDocumento            *string        `json:"NumeroDocumento,omitempty"`            // String(20)	*Número do documento do fornecedor.
	FormaPagamento             *string        `json:"FormaPagamento,omitempty"`             // *Forma de pagamento do lançamento.
	TipoDocumento              *string        `json:"TipoDocumento,omitempty"`              // *Tipo de documento do lançamento.
	NFSE                       *string        `json:"NFSE,omitempty"`                       // Indica se o documento é nota fiscal eletrônica. Valor default é 'N'.
	Complemento                *string        `json:"Complemento,omitempty"`                // Complemento descritivo do lançamento.
	ComplementoAdicional1      *string        `json:"ComplementoAdicional1,omitempty"`      // Informação de complemento extra.
	ComplementoAdicional2      *string        `json:"ComplementoAdicional2,omitempty"`      // Informação de complemento extra.
	ComplementoAdicional3      *string        `json:"ComplementoAdicional3,omitempty"`      // Informação de complemento extra.
	ComplementoAdicional4      *string        `json:"ComplementoAdicional4,omitempty"`      // Informação de complemento extra.
	ComplementoAdicional5      *string        `json:"ComplementoAdicional5,omitempty"`      // Informação de complemento extra.
	ComplementoAdicional6      *string        `json:"ComplementoAdicional6,omitempty"`      // Informação de complemento extra.
	ComplementoAdicional7      *string        `json:"ComplementoAdicional7,omitempty"`      // Informação de complemento extra.
	ComplementoAdicional8      *string        `json:"ComplementoAdicional8,omitempty"`      // Informação de complemento extra.
	NumeroParcela              *int           `json:"NumeroParcela,omitempty"`              // Número da parcela do lançamento. Valor default é '1'.
	TotalParcelas              *int           `json:"TotalParcelas,omitempty"`              // Quantidade total de parcelas. Valor default é '1'.
	ContaCorrente              *string        `json:"ContaCorrente,omitempty"`              // Número da conta corrente da qual originará o pagamento bancário quando aplicado.
	CodigoBarras               *string        `json:"CodigoBarras,omitempty"`               // Código de barras do documento (* obrigatório se origem for 'B')
	PixQrCode                  *string        `json:"PixQrCode,omitempty"`                  // QR Code.
	DataEmissao                *string        `json:"DataEmissao,omitempty"`                // Data de emissão do lançamento (se TipoDocumento for 'N').
	DataVencimento             *string        `json:"DataVencimento,omitempty"`             // *Data de vencimento do lançamento.
	PrevisaoReal               *string        `json:"PrevisaoReal,omitempty"`               // *Indicação de lançamento previsto ou real.
	Frequencia                 *string        `json:"Frequencia,omitempty"`                 // Define se lançamento é único ou permanente. Valor default é 'U'.
	ValorTotal                 *types.Float64 `json:"ValorTotal,omitempty"`                 // Valor total do documento. Quando lançamento é uma parcela, informar o valor bruto do parcelamento. Caso não seja parcelamento este campo será ignorado.
	ValorBruto                 *types.Float64 `json:"ValorBruto,omitempty"`                 // *Valor bruto do documento/parcela.
	ValorDescontoIncondicional *types.Float64 `json:"ValorDescontoIncondicional,omitempty"` // Valor do desconto incondicional. Este desconto é abatido da base de cálculo de impostos.
	ValorDescontoCondicional   *types.Float64 `json:"ValorDescontoCondicional,omitempty"`   // Valor do desconto condicional. Este desconto não é abatido da base de cálculo de impostos.
	ValorJuros                 *types.Float64 `json:"ValorJuros,omitempty"`                 // Valor dos juros.
	ValorServicos              *types.Float64 `json:"ValorServicos,omitempty"`              // Valor dos serviços. Se não informado, a base de cálculo será ValorBruto.
	ValorBaseCalculoIss        *types.Float64 `json:"ValorBaseCalculoIss,omitempty"`        // Base de cálculo do ISS. Se não informado, a base de cálculo será ValorServicos.
	ValorRetencaoInss          *types.Float64 `json:"ValorRetencaoInss,omitempty"`          // Valor do INSS a ser retido.
	ValorRetencaoIss           *types.Float64 `json:"ValorRetencaoIss,omitempty"`           // Valor do ISS a ser retido.
	ValorRetencaoIrf           *types.Float64 `json:"ValorRetencaoIrf,omitempty"`           // Valor do IRF a ser retido.
	ValorRetencaoFederal       *types.Float64 `json:"ValorRetencaoFederal,omitempty"`       // Valor da retenção federal a ser retida.
	NomePagador                *string        `json:"NomePagador,omitempty"`                // Nome do beneficiário. (Para liquidação de títulos se este for diferente do condomínio).
	TipoPessoaPagador          *string        `json:"TipoPessoaPagador,omitempty"`          // Tipo de pessoa do pagador. (Para liquidação de títulos se este for diferente do condomínio).
	CpfCnpjPagador             *string        `json:"CpfCnpjPagador,omitempty"`             // CPF ou CNPJ do pagador. (Para liquidação de títulos se este for diferente do condomínio).
	NomeBeneficiario           *string        `json:"NomeBeneficiario,omitempty"`           // Nome do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	TipoPessoaBeneficiario     *string        `json:"TipoPessoaBeneficiario,omitempty"`     // Tipo de pessoa do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	CpfCnpjBeneficiario        *string        `json:"CpfCnpjBeneficiario,omitempty"`        // CPF ou CNPJ do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	GrupoSoma                  *int           `json:"GrupoSoma,omitempty"`                  // Código do grupo de soma.
	CodigoImagem               *string        `json:"CodigoImagem,omitempty"`               // Código da imagem do lançamento.
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
