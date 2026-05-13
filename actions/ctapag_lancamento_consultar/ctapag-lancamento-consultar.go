package ctapag_lancamento_consultar

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

var ACTION = "CTAPAG_LANCAMENTO_CONSULTAR"

type ActionInput struct {
	NumeroLancto *int `json:"NumeroLancto,omitempty"` // *Número do lançamento.
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
	NumeroLancto           *int                           `json:"NumeroLancto,omitempty"`           // Número do lançamento.
	Origem                 *string                        `json:"Origem,omitempty"`                 // Área de origem do lançamento.
	CodFilial              *string                        `json:"CodFilial,omitempty"`              // Código da filial do lançamento.
	CodCondominio          *int                           `json:"CodCondominio,omitempty"`          // Código do condomínio do lançamento (se origem for 'C').
	NomeCondominio         *string                        `json:"NomeCondominio,omitempty"`         // Nome do condomínio do lançamento (se origem for 'C').
	CodBloco               *string                        `json:"CodBloco,omitempty"`               // Código do bloco do lançamento (se origem for 'C').
	CodImovel              *int                           `json:"CodImovel,omitempty"`              // Código do imóvel do lançamento (se origem for 'I').
	EnderecoImovel         *string                        `json:"EnderecoImovel,omitempty"`         // Endereço do imóvel (se origem for 'I').
	CodLocatario           *int                           `json:"CodLocatario,omitempty"`           // Código de pessoa do locatário (se origem for 'I').
	NomeLocatario          *string                        `json:"NomeLocatario,omitempty"`          // Nome do locatário (se origem for 'I').
	CodPessoaProprietario  *int                           `json:"CodPessoaProprietario,omitempty"`  // Código de pessoa do proprietário (se origem for 'R').
	NomeProprietario       *string                        `json:"NomeProprietario,omitempty"`       // Nome do proprietário (se origem for 'R').
	CodPlanoContaAdm       *int                           `json:"CodPlanoContaAdm,omitempty"`       // Código da conta no plano de contas da administradora (se origem for 'A').
	DescrPlanoContaAdm     *int                           `json:"DescrPlanoContaAdm,omitempty"`     // Descrição da conta no plano de contas da administradora (se origem for 'A').
	CodFornecedor          *int                           `json:"CodFornecedor,omitempty"`          // Código do fornecedor do lançamento.
	NomeFornecedor         *string                        `json:"NomeFornecedor,omitempty"`         // Nome do fornecedor do lançamento.
	CodPessoaFavorecido    *int                           `json:"CodPessoaFavorecido,omitempty"`    // C´dogido da pessoa informada como favorecido.
	NomeFavorecido         *string                        `json:"NomeFavorecido,omitempty"`         // Nome do favorecido.
	Competencia            *string                        `json:"Competencia,omitempty"`            // Competência do lançamento no formato 'YYYYMM'.
	DataEmissao            *string                        `json:"DataEmissao,omitempty"`            // Data de emissão do lançamento (se TipoDocumento for 'N').
	DataVencimento         *string                        `json:"DataVencimento,omitempty"`         // Data de vencimento do lançamento.
	DataPagamento          *string                        `json:"DataPagamento,omitempty"`          // Data de pagamento do lançamento (quando quitado).
	FormaPagamento         *string                        `json:"FormaPagamento,omitempty"`         // Forma de pagamento do lançamento.
	TipoDocumento          *string                        `json:"TipoDocumento,omitempty"`          // Tipo de documento do lançamento.
	NFSE                   *string                        `json:"NFSE,omitempty"`                   // Indica se o documento é nota fiscal eletrônica.
	CodTaxa                *int                           `json:"CodTaxa,omitempty"`                // Código da taxa que classifica este lançamento.
	DescrTaxa              *string                        `json:"DescrTaxa,omitempty"`              // Descrição da taxa que classifica este lançamento.
	NumeroParcela          *int                           `json:"NumeroParcela,omitempty"`          // Número da parcela do lançamento.
	TotalParcelas          *int                           `json:"TotalParcelas,omitempty"`          // Quantidade total de parcelas.
	Complemento            *string                        `json:"Complemento,omitempty"`            // Complemento descritivo do lançamento.
	NumeroDocumento        *string                        `json:"NumeroDocumento,omitempty"`        // Número do documento do fornecedor.
	ValorBruto             *types.Float64                       `json:"ValorBruto,omitempty"`             // Valor bruto do documento/parcela.
	ValorServicos          *types.Float64                       `json:"ValorServicos,omitempty"`          // Valor dos serviços. Se não informado, a base de cálculo será ValorBruto.
	ValorBaseCalculoIss    *types.Float64                       `json:"ValorBaseCalculoIss,omitempty"`    // Base de cálculo do ISS. Se não informado, a base de cálculo será ValorServicos.
	ValorRetencaoInss      *types.Float64                       `json:"ValorRetencaoInss,omitempty"`      // Valor do INSS a ser retido.
	ValorRetencaoIss       *types.Float64                       `json:"ValorRetencaoIss,omitempty"`       // Valor do ISS a ser retido.
	ValorRetencaoIrf       *types.Float64                       `json:"ValorRetencaoIrf,omitempty"`       // Valor do IRF a ser retido.
	ValorRetencaoFederal   *types.Float64                       `json:"ValorRetencaoFederal,omitempty"`   // Valor da retenção federal a ser retida.
	ValorDesconto          *types.Float64                       `json:"ValorDesconto,omitempty"`          // Valor do desconto.
	ValorJuros             *types.Float64                       `json:"ValorJuros,omitempty"`             // Valor dos juros.
	Comissao               *types.Float64                       `json:"Comissao,omitempty"`               // Valor de comissão.
	CodigoBarras           *string                        `json:"CodigoBarras,omitempty"`           // Código de barras do documento (* obrigatório se origem for 'B')
	PrevisaoReal           *string                        `json:"PrevisaoReal,omitempty"`           // Indicação de lançamento previsto ou real.
	Frequencia             *string                        `json:"Frequencia,omitempty"`             // Define se lançamento é único ou permanente.
	UsuarioSuspensao       *string                        `json:"UsuarioSuspensao,omitempty"`       // Usuário que suspendeu o lançamento.
	DataSuspensao          *string                        `json:"DataSuspensao,omitempty"`          // Data da suspensão do lançamento.
	MotivoSuspensao        *string                        `json:"MotivoSuspensao,omitempty"`        // Motivo da suspensão do lançamento.
	CodPessoaBeneficiario  *int                           `json:"CodPessoaBeneficiario,omitempty"`  // Código da pessoa definida como beneficiário do pagamento.
	NomeBeneficiario       *string                        `json:"NomeBeneficiario,omitempty"`       // Nome do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	TipoPessoaBeneficiario *string                        `json:"TipoPessoaBeneficiario,omitempty"` // Tipo de pessoa do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	CpfCnpjBeneficiario    *int                           `json:"CpfCnpjBeneficiario,omitempty"`    // CPF ou CNPJ do beneficiário. (Para liquidação de títulos se este for diferente do fornecedor/favorecido).
	CodPessoaPagador       *int                           `json:"CodPessoaPagador,omitempty"`       // Código do pagador no cadastro de pessoas.
	NomePagador            *string                        `json:"NomePagador,omitempty"`            // Nome do beneficiário. (Para liquidação de títulos se este for diferente do condomínio).
	TipoPessoaPagador      *string                        `json:"TipoPessoaPagador,omitempty"`      // Tipo de pessoa do pagador. (Para liquidação de títulos se este for diferente do condomínio).
	CpfCnpjPagador         *int                           `json:"CpfCnpjPagador,omitempty"`         // CPF ou CNPJ do pagador. (Para liquidação de títulos se este for diferente do condomínio).
	Agrupados              *[]RequestResponseBodyAgrupado `json:"Agrupados,omitempty"`              // Lista de lançamentos agrupados no lançamento agrupador.
}

type RequestResponseBodyAgrupado struct {
	NumeroLancto *int     `json:"NumeroLancto,omitempty"` // Number(10)	Número do lançamento.
	Origem       *string  `json:"Origem,omitempty"`       // String(1)	Área de origem do lançamento.
	CodigoOrigem *string  `json:"CodigoOrigem,omitempty"` // String(10)	Código do condomínio ou imóvel ou pessoa ou conta da administradora.
	CodTaxa      *int     `json:"CodTaxa,omitempty"`      // Number(5)	Código da taxa que classifica este lançamento.
	Valor        *types.Float64 `json:"Valor,omitempty"`        // Number(12,2)	Valor do lançamento.
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
