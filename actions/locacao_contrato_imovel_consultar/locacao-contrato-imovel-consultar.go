package locacao_contrato_imovel_consultar

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

var ACTION = "LOCACAO_CONTRATO_IMOVEL_CONSULTAR"

type ActionInput struct {
	CodImovel      *int `json:"CodImovel,omitempty"`      // *Código do imóvel.
	CodContratoLoc *int `json:"CodContratoLoc,omitempty"` // Código do contrato de locação deste imóvel.
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
	CodImovel                 *int                            `json:"CodImovel,omitempty"`                 // Código do imóvel.
	CodContratoLoc            *int                            `json:"CodContratoLoc,omitempty"`            // Código do contrato de locação deste imóvel.
	CodIntegracaoSist         *string                         `json:"CodIntegracaoSist,omitempty"`         // Código deste contrato de locação no sistema integrado/migrado.
	CodContratoAdm            *int                            `json:"CodContratoAdm,omitempty"`            // Código do contrato de administração deste imóvel.
	DataAssinatura            *string                         `json:"DataAssinatura,omitempty"`            // Data de assinatura do contrato.
	DataDistrato              *string                         `json:"DataDistrato,omitempty"`              // Data distrato.
	ContratoLocAtivo          *string                         `json:"ContratoLoc_Ativo,omitempty"`         // Indica se o contrato de locação está ativo.
	CodPessoaLocat            *int                            `json:"CodPessoaLocat,omitempty"`            // Código de pessoa do locatário principal.
	DataVigInicial            *string                         `json:"DataVigInicial,omitempty"`            // Data inicial da vigência do contrato.
	DataVigFinal              *string                         `json:"DataVigFinal,omitempty"`              // Data final da vigência do contrato.
	DataProxReaj              *string                         `json:"DataProxReaj,omitempty"`              // Data do próximo reajuste.
	DataEntregaChaves         *string                         `json:"DataEntregaChaves,omitempty"`         // Data da entrega de chaves na desocupação.
	DataAvisoDesocupacao      *string                         `json:"DataAvisoDesocupacao,omitempty"`      // Data do aviso de desocupação.
	ValorAluguel              *float64                        `json:"ValorAluguel,omitempty"`              // Valor do aluguel inicial.
	DataVencFianca            *string                         `json:"DataVencFianca,omitempty"`            // Data de vencimento do seguro fiança.
	DescricaoUso              *string                         `json:"DescricaoUso,omitempty"`              // Descrição de qual será a utilização do imóvel.
	CodAgencia                *int                            `json:"CodAgencia,omitempty"`                // Código da agência/loja a qual este contrato pertence.
	PeriodoReajAluguel        *int                            `json:"PeriodoReajAluguel,omitempty"`        // Indica o número de meses em que o aluguel sofrerá reajuste, ou seja, de quantos em quantos meses ele será reajustado.
	Prazo                     *int                            `json:"Prazo,omitempty"`                     // Prazo do contrato em meses.
	PercMulta                 *float64                        `json:"PercMulta,omitempty"`                 // Percentual de multa em caso de atraso de pagamento.
	PercJuros                 *float64                        `json:"PercJuros,omitempty"`                 // Percentual de juros em caso de atraso de pagamento.
	Carencia                  *string                         `json:"Carencia,omitempty"`                  // Indica se o locatário possui um período de carência inicial.
	MesesCarencia             *int                            `json:"MesesCarencia,omitempty"`             // Indica quantos meses de carência é dado ao locatário.
	DiasCarencia              *int                            `json:"DiasCarencia,omitempty"`              // Indica quantos dias de carência é dado ao locatário. Se for em percentual então informar no campo 'PercCarencia' mas apenas um deles deve ser informado.
	PercCarencia              *float64                        `json:"PercCarencia,omitempty"`              // Indica o percentual de carência que é dado ao locatário dentro do mês. Se for em número de dias então informar no campo 'DiasCarencia' mas apenas um deles deve ser informado.
	DescPontualidade          *string                         `json:"DescPontualidade,omitempty"`          // Indica se tem desconto de pontualidade quando pago antes do vencimento.
	PercDescPontualidade      *float64                        `json:"PercDescPontualidade,omitempty"`      // Percentual de desconto pontualidade.
	ValorDescPontualidade     *float64                        `json:"ValorDescPontualidade,omitempty"`     // Valor fixo em Reais de desconto pontualidade, caso não se utilize um percentual de desconto.
	DiasDescPontualidade      *int                            `json:"DiasDescPontualidade,omitempty"`      // Número mínimo de dias de antecipação do pagamento para habilitar o desconto pontualidade.
	FormaCalcPagto            *string                         `json:"FormaCalcPagto,omitempty"`            // Indica a forma de cálculo para o pagamento ao proprietário.
	DiaPagtoProp              *int                            `json:"DiaPagtoProp,omitempty"`              // Dia do mês para o pagamento ao proprietário quando a forma de cálculo for 'Programado'.
	DiaVenctoDOC              *int                            `json:"DiaVenctoDOC,omitempty"`              // Dia do mês para vencimento do DOC de aluguel.
	IsentaIR                  *string                         `json:"IsentaIR,omitempty"`                  // Indica se o aluguel é isento de imposto de renda.
	IsentaDescontoValorBaseIR *string                         `json:"IsentaDescontoValorBaseIR,omitempty"` // Indica se é isento do desconto do valor base do imposto de renda.
	DocPorPeriodo             *string                         `json:"DocPorPeriodo,omitempty"`             // Indica se o aluguel é por período determinado.
	LocacaoTemporada          *string                         `json:"Locacao_Temporada,omitempty"`         // Indica se o aluguel é por temporada.
	IsentaTaxaPorte           *string                         `json:"IsentaTaxaPorte,omitempty"`           // Indica se deve isentar da taxa porte.
	IsentaTarifaDOC           *string                         `json:"IsentaTarifaDOC,omitempty"`           // Indica se deve isentar da tarifa DOC.
	DOCEmail                  *string                         `json:"DOC_Email,omitempty"`                 // Indica se o DOC deve ser enviado por E-mail.
	GaranteDOC                *string                         `json:"GaranteDOC,omitempty"`                // Indica se utiliza a modalidade de DOC garantido.
	ValorTarifaDOC            *float64                        `json:"ValorTarifaDOC,omitempty"`            // Valor da tarifa DOC, caso não seja o valor default do sistema.
	PercReajAluguel           *float64                        `json:"PercReajAluguel,omitempty"`           // Percentual para correção do valor de aluguel, caso não se utilize um índice de reajuste.
	ValorTxIntermed           *float64                        `json:"ValorTxIntermed,omitempty"`           // Valor da taxa de intermediação.
	CompetIniIntermed         *string                         `json:"CompetIniIntermed,omitempty"`         // Competencia da cobrança inicial da taxa de intermediação no formato 'YYYYMM'.
	NrParcIntermed            *int                            `json:"NrParcIntermed,omitempty"`            // Número de parcelas para pagamento da intermediação.
	TipoAditamento            *string                         `json:"TipoAditamento,omitempty"`            // Tipo do aditamento.
	IndiceReajAluguel         *string                         `json:"IndiceReajAluguel,omitempty"`         // Índice monetário para correção do valor de aluguel.
	TipoFianca                *string                         `json:"TipoFianca,omitempty"`                // Tipo de seguro fiança.
	NomeLocat                 *string                         `json:"NomeLocat,omitempty"`                 // Nome do locatário principal.
	TaxaAdm                   *float64                        `json:"TaxaAdm,omitempty"`                   // Taxa de administração do imóvel em forma de um percentual sobre o aluguel. Se for um valor fixo em Reais então informá-lo no campo 'ValorTaxaAdm' mas apenas um deles deve ser informado.
	ValorTaxaAdm              *float64                        `json:"ValorTaxaAdm,omitempty"`              // Taxa de administração do imóvel em forma de um valor fixo em Reais. Se for um percentual sobre o aluguel então informá-lo no campo 'TaxaAdm' mas apenas um deles deve ser informado.
	TipoAssinatura            *string                         `json:"TipoAssinatura,omitempty"`            // Tipo de assinatura do contrato de locação.
	CodFornecCobr             *int                            `json:"CodFornecCobr,omitempty"`             // Código de fornecedor do escritório de cobrança.
	CodFornecFianca           *int                            `json:"CodFornecFianca,omitempty"`           // Código de fornecedor da seguradora do seguro fiança.
	HonorariosPerc            *float64                        `json:"HonorariosPerc,omitempty"`            // Percentual de honorários.
	HonorariosPrazoDias       *int                            `json:"HonorariosPrazoDias,omitempty"`       // Prazo em dias de honorários.
	Fiadores                  *[]RequestResponseBodyFiador    `json:"Fiadores,omitempty"`                  // Lista de pessoas que são fiadores.
	Locatarios                *[]RequestResponseBodyLocatario `json:"Locatarios,omitempty"`                // Lista de pessoas que são locatários adicionais.
	Agrupados                 *[]RequestResponseBodyAgrupado  `json:"Agrupados,omitempty"`                 //
}

type RequestResponseBodyFiador struct {
	CodPessoa *int `json:"CodPessoa,omitempty"` // Código de pessoa.
}

type RequestResponseBodyLocatario struct {
	CodPessoa *int `json:"CodPessoa,omitempty"` // Código de pessoa.
}

type RequestResponseBodyAgrupado struct {
	CodImovel      *int    `json:"CodImovel,omitempty"`      // Código do imóvel.
	CodContratoLoc *int    `json:"CodContratoLoc,omitempty"` // Código do contrato de locação deste imóvel.
	Endereco       *string `json:"Endereco,omitempty"`       // Endereço do imóvel.
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
