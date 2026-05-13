package locacao_relatorio_mensal

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

var ACTION = "LOCACAO_RELATORIO_MENSAL"

type ActionInput struct {
	Competencia       *string `json:"Competencia,omitempty"`       // *Competência do relatório mensal a gerar.
	CodFilial         *int    `json:"CodFilial,omitempty"`         // Código da filial a gerar. Valor default é '000'.
	DadosBoletos      *string `json:"DadosBoletos,omitempty"`      // Indica para gerar dados totalizados dos boletos. Valor default é 'S'.
	TaxaBoletos       *string `json:"TaxaBoletos,omitempty"`       // Indica para gerar relação de taxas dos boletos. Valor default é 'N'.
	TaxaAdministracao *string `json:"TaxaAdministracao,omitempty"` // Indica para gerar relação de taxas de administração. Valor default é 'N'.
	PagtoProprietario *string `json:"PagtoProprietario,omitempty"` // Indica para gerar relação de pagamento de proprietário. Valor default é 'N'.
	NumerosLocacao    *string `json:"NumerosLocacao,omitempty"`    // Indica para gerar relação de números da locação. Valor default é 'N'.
	BoletosBancos     *string `json:"BoletosBancos,omitempty"`     // Indica para gerar informações sintéticas dos boletos por banco. Valor default é 'N'.
	ResponseFormat    *string `json:"ResponseFormat,omitempty"`    // Formato desejado da resposta.
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
	TiposImoveis             *[]RequestResponseBodyTipoImovel             //
	Competencia              *string                                      // Competência do relatório mensal a gerar.
	TaxasBoletos             *[]RequestResponseBodyTaxaBoleto             //
	TotaisTaxas              *RequestResponseBodyTotaisTaxas              //
	DadosBoletos             *RequestResponseBodyDadosBoletos             // Indica para gerar dados totalizados dos boletos.
	QuadroBoletos            *RequestResponseBodyQuadroBoletos            //
	Inadimplencias           *RequestResponseBodyInadimplencias           //
	TaxasAdministracao       *RequestResponseBodyTaxasAdministracao       //
	QuadroPagtoProprietarios *RequestResponseBodyQuadroPagtoProprietarios //
	NumerosLocacao           *RequestResponseBodyNumerosLocacao           // Indica para gerar relação de números da locação.
	SituacoesImoveis         *[]RequestResponseBodySituacaoImoveis        //
	TotaisLocacao            *RequestResponseBodyTotaisLocacao            //
	TiposBoletos             *[]RequestResponseBodyTipoBoletos            //
}

type RequestResponseBodyTipoImovel struct {
	Titulo  *string                                `json:"Titulo,omitempty"`  // Tipos de imóveis agrupados.
	Imoveis *[]RequestResponseBodyTipoImovelImovel `json:"Imoveis,omitempty"` //
}

type RequestResponseBodyTipoImovelImovel struct {
	Classe    *string  `json:"Classe,omitempty"`    // Classe do imóvel.
	QtdClasse *int     `json:"QtdClasse,omitempty"` // Quantidade de imóveis desta classe.
	VlrClasse *types.Float64 `json:"VlrClasse,omitempty"` // Valor total de aluguéis desta classe.
}

type RequestResponseBodyTaxaBoleto struct {
	CodTaxa *int     `json:"CodTaxa,omitempty"` // Codigo da taxa.
	Taxa    *string  `json:"Taxa,omitempty"`    // Descrição da taxa.
	Debito  *types.Float64 `json:"Debito,omitempty"`  // Valor de débito.
	Credito *types.Float64 `json:"Credito,omitempty"` // Valor de crédito.
	Total   *types.Float64 `json:"Total,omitempty"`   // Valor total.
}

type RequestResponseBodyTotaisTaxas struct {
	Debito  *types.Float64 `json:"Debito,omitempty"`  // Valor de débito.
	Credito *types.Float64 `json:"Credito,omitempty"` // Valor de crédito.
	Total   *types.Float64 `json:"Total,omitempty"`   // Valor total.
}

type RequestResponseBodyDadosBoletos struct {
	TaxaPorte           *types.Float64 `json:"TaxaPorte,omitempty"`           // Valor de taxa porte.
	DescAdministradora  *types.Float64 `json:"DescAdministradora,omitempty"`  // Valor de desconto da administradora.
	MultaAdministradora *types.Float64 `json:"MultaAdministradora,omitempty"` // Valor de multa da administradora.
	JurosAdministradora *types.Float64 `json:"JurosAdministradora,omitempty"` // Valor de juros da administradora.
	OutrosAcrescimos    *types.Float64 `json:"OutrosAcrescimos,omitempty"`    // Valor de outros acréscimos.
	TarifaDoc           *types.Float64 `json:"TarifaDoc,omitempty"`           // Valor de emissão do boleto (tarifa DOC).
	DescProprietario    *types.Float64 `json:"DescProprietario,omitempty"`    // Valor de desconto do proprietário.
	MultaProprietario   *types.Float64 `json:"MultaProprietario,omitempty"`   // Valor de multa do proprietário.
	JurosProprietario   *types.Float64 `json:"JurosProprietario,omitempty"`   // Valor de juros do proprietário.
}

type RequestResponseBodyQuadroBoletos struct {
	VlrBoletosEmitidos            *types.Float64 `json:"VlrBoletosEmitidos,omitempty"`            // Valor total de boletos emitidos.
	QtdBoletosEmitidos            *int     `json:"QtdBoletosEmitidos,omitempty"`            // Quantidade de boletos emitidos.
	VlrBoletosPagosMesAtual       *types.Float64 `json:"VlrBoletosPagosMesAtual,omitempty"`       // Valor total de boletos pagos do mês.
	QtdBoletosPagosMesAtual       *int     `json:"QtdBoletosPagosMesAtual,omitempty"`       // Quantidade de boletos pagos do mês.
	VlrBoletosPagosMesAnt         *types.Float64 `json:"VlrBoletosPagosMesAnt,omitempty"`         // Valor total de boletos pagos meses anteriores.
	QtdBoletosPagosMesAnt         *int     `json:"QtdBoletosPagosMesAnt,omitempty"`         // Quantidade de boletos pagos meses anteriores.
	VlrBoletosPagosMesFuturo      *types.Float64 `json:"VlrBoletosPagosMesFuturo,omitempty"`      // Valor total de boletos pagos meses futuros.
	QtdBoletosPagosMesFuturo      *int     `json:"QtdBoletosPagosMesFuturo,omitempty"`      // Quantidade de boletos pagos meses futuros.
	VlrBoletosNaoPagosCompetAtual *types.Float64 `json:"VlrBoletosNaoPagosCompetAtual,omitempty"` // Valor total de boletos não quitados na competência.
	QtdBoletosNaoPagosCompetAtual *int     `json:"QtdBoletosNaoPagosCompetAtual,omitempty"` // Quantidade de boletos não quitados na competência.
	VlrBoletosNaoPagosCompetAnt   *types.Float64 `json:"VlrBoletosNaoPagosCompetAnt,omitempty"`   // Valor total de boletos não quitados em competências anteriores.
	QtdBoletosNaoPagosCompetAnt   *int     `json:"QtdBoletosNaoPagosCompetAnt,omitempty"`   // Quantidade de boletos não quitados em competências anteriores.
}

type RequestResponseBodyInadimplencias struct {
	VlrInadimplencias *types.Float64 `json:"VlrInadimplencias,omitempty"` // Valor total de boletos de aluguéis atrasados.
	QtdInadimplencias *int     `json:"QtdInadimplencias,omitempty"` // Quantidade de boletos de aluguéis atrasados.
}

type RequestResponseBodyTaxasAdministracao struct {
	VlrPrevisaoBoletosMes     *types.Float64 `json:"VlrPrevisaoBoletosMes,omitempty"`     // Valor da previsão de taxa de administração dos boletos gerados do mês.
	VlrEfetivoBoletosMes      *types.Float64 `json:"VlrEfetivoBoletosMes,omitempty"`      // Valor total da taxa de administração efetiva dos boletos do mês.
	VlrTotalRecebidoMes       *types.Float64 `json:"VlrTotalRecebidoMes,omitempty"`       // Valor total da taxa de administração recebida no mês.
	VlrRecebidoMesSemGarantia *types.Float64 `json:"VlrRecebidoMesSemGarantia,omitempty"` // Valor da taxa de administração recebida no mês - sem garantia.
	VlrGarantidoMes           *types.Float64 `json:"VlrGarantidoMes,omitempty"`           // Valor da taxa de administração garantida no mês.
	VlrGarantidoRecebidoMes   *types.Float64 `json:"VlrGarantidoRecebidoMes,omitempty"`   // Valor da taxa de administração garantida e recebida no mesmo mês.
	VlrGarantidoOutroMes      *types.Float64 `json:"VlrGarantidoOutroMes,omitempty"`      // Valor da taxa de administração de doc garantido em outro mês e quitado neste.
	VlrDemonstrativoProp      *types.Float64 `json:"VlrDemonstrativoProp,omitempty"`      // Valor da taxa de admnistração dentro dos demonstrativos de proprietários.
	VlrIntermediacaoMes       *types.Float64 `json:"VlrIntermediacaoMes,omitempty"`       // Valor total da taxa de intermediação efetiva no mês.
}

type RequestResponseBodyQuadroPagtoProprietarios struct {
	VlrGerado        *types.Float64 `json:"VlrGerado,omitempty"`        // Valor total de gerações no mês.
	QtdGerado        *int     `json:"QtdGerado,omitempty"`        // Quantidade de gerações no mês.
	VlrPago          *types.Float64 `json:"VlrPago,omitempty"`          // Valor total de gerações pagas.
	QtdPago          *int     `json:"QtdPago,omitempty"`          // Quantidade de gerações pagas.
	VlrTarifaRemessa *types.Float64 `json:"VlrTarifaRemessa,omitempty"` // Valor total de tarifa remessa.
	QtdTarifaRemessa *int     `json:"QtdTarifaRemessa,omitempty"` // Quantidade de tarifa remessa.
}

type RequestResponseBodyNumerosLocacao struct {
	QtdProprietarios               *int `json:"QtdProprietarios,omitempty"`               // Quantidade de proprietários.
	QtdTotalImoveis                *int `json:"QtdTotalImoveis,omitempty"`                // Quantidade total de imóveis.
	QtdOcupados                    *int `json:"QtdOcupados,omitempty"`                    // Quantidade de imóveis ocupados.
	QtdOcupadosEmVenda             *int `json:"QtdOcupadosEmVenda,omitempty"`             // Quantidade de imóveis ocupados em venda.
	QtdCortesia                    *int `json:"QtdCortesia,omitempty"`                    // Quantidade de imóveis cortesia.
	QtdDesocupados                 *int `json:"QtdDesocupados,omitempty"`                 // Quantidade de imóveis desocupados.
	QtdDesocupComComercializacao   *int `json:"QtdDesocupComComercializacao,omitempty"`   // Quantidade de imóveis desocupados com comercialização.
	QtdDesocupLocacao              *int `json:"QtdDesocupLocacao,omitempty"`              // Quantidade de imóveis desocupados com comercialização para locação.
	QtdDesocupVenda                *int `json:"QtdDesocupVenda,omitempty"`                // Quantidade de imóveis desocupados com comercialização para venda.
	QtdDesocupVendaLocacao         *int `json:"QtdDesocupVendaLocacao,omitempty"`         // Quantidade de imóveis desocupados com comercialização para venda e locação.
	QtdDesocupSemComercializacao   *int `json:"QtdDesocupSemComercializacao,omitempty"`   // Quantidade de imóveis desocupados sem comercialização.
	QtdDesocupSemComercComContrato *int `json:"QtdDesocupSemComercComContrato,omitempty"` // Quantidade de imóveis desocupados sem comercialização e com contrato de administração.
	QtdInquilinos                  *int `json:"QtdInquilinos,omitempty"`                  // Quantidade de inquilinos.
}

type RequestResponseBodySituacaoImoveis struct {
	Situacao   *string `json:"Situacao,omitempty"`   // Situação dos imóveis.
	Quantidade *int    `json:"Quantidade,omitempty"` // Quantidade de imóveis nesta situação.
}

type RequestResponseBodyTotaisLocacao struct {
	QtdTotalImoveis         *int     `json:"QtdTotalImoveis,omitempty"`         // Quantidade total de imóveis.
	VlrAlugueisResidenciais *types.Float64 `json:"VlrAlugueisResidenciais,omitempty"` // Valor total de aluguel para imóveis residenciais.
	QtdAlugueisResidenciais *int     `json:"QtdAlugueisResidenciais,omitempty"` // Quantidade de imóveis com aluguel residencial.
	VlrAlugueisComerciais   *types.Float64 `json:"VlrAlugueisComerciais,omitempty"`   // Valor total de aluguel para imóveis comerciais.
	QtdAlugueisComerciais   *int     `json:"QtdAlugueisComerciais,omitempty"`   // Quantidade de imóveis com aluguel comercial.
	VlrAlugueis             *types.Float64 `json:"VlrAlugueis,omitempty"`             // Valor total de aluguéis.
}

type RequestResponseBodyTipoBoletos struct {
	Descricao *string                                `json:"Descricao,omitempty"` // Tipos de boletos.
	Bancos    *[]RequestResponseBodyTipoBoletosBanco //
}

type RequestResponseBodyTipoBoletosBanco struct {
	CodBanco    *int                                            `json:"CodBanco,omitempty"`    // Código do banco.
	NomeBanco   *string                                         `json:"NomeBanco,omitempty"`   // Nome do banco.
	Boletos     *[]RequestResponseBodyTipoBoletosBancoBoleto    `json:"Boletos,omitempty"`     //
	TotaisBanco *RequestResponseBodyTipoBoletosBancoTotaisBanco `json:"TotaisBanco,omitempty"` //
}

type RequestResponseBodyTipoBoletosBancoBoleto struct {
	Data      *string  `json:"Data,omitempty"`      // Data.
	QtdTotal  *int     `json:"QtdTotal,omitempty"`  // Quantidade de boletos emitidos.
	VlrTotal  *types.Float64 `json:"VlrTotal,omitempty"`  // Valor total de boletos emitidos.
	QtdNormal *int     `json:"QtdNormal,omitempty"` // Quantidade de boletos normais/extras.
	VlrNormal *types.Float64 `json:"VlrNormal,omitempty"` // Valor dos boletos normais/extras.
	QtdRetido *int     `json:"QtdRetido,omitempty"` // Quantidades de boletos rettidos.
	VlrRetido *types.Float64 `json:"VlrRetido,omitempty"` // Valor dos boletos retidos.
}

type RequestResponseBodyTipoBoletosBancoTotaisBanco struct {
	QtdTotal  *int     `json:"QtdTotal,omitempty"`  // Quantidade de boletos emitidos.
	VlrTotal  *types.Float64 `json:"VlrTotal,omitempty"`  // Valor total de boletos emitidos.
	QtdNormal *int     `json:"QtdNormal,omitempty"` // Quantidade de boletos normais/extras.
	VlrNormal *types.Float64 `json:"VlrNormal,omitempty"` // Valor dos boletos normais/extras.
	QtdRetido *int     `json:"QtdRetido,omitempty"` // Quantidades de boletos rettidos.
	VlrRetido *types.Float64 `json:"VlrRetido,omitempty"` // Valor dos boletos retidos.
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
