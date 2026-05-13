package condom_relatorio_mensal

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

var ACTION = "CONDOM_RELATORIO_MENSAL"

type ActionInput struct {
	Competencia    *string `json:"Competencia,omitempty"`    // *Competência do relatório mensal a gerar.
	CodFilial      *int    `json:"CodFilial,omitempty"`      // Código da filial a gerar. Valor default é '000'.
	InfosExtras    *string `json:"InfosExtras,omitempty"`    // Indica para gerar informações extras. Valor default é 'N'.
	BoletosBancos  *string `json:"BoletosBancos,omitempty"`  // Indica para gerar informações sintéticas dos boletos por banco. Valor default é 'N'.
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
	Competencia  *string                           `json:"Competencia,omitempty"`  // Competência do relatório mensal a gerar.
	InfosGerais  *RequestResponseBodyInfosGerais   `json:"InfosGerais,omitempty"`  //
	InfosExtras  *RequestResponseBodyInfosExtras   `json:"InfosExtras,omitempty"`  // Indica para gerar informações extras.
	TiposBoletos *[]RequestResponseBodyTipoBoletos `json:"TiposBoletos,omitempty"` //
}

type RequestResponseBodyInfosGerais struct {
	QtdCondomAtivos        *int     `json:"QtdCondomAtivos,omitempty"`        // Quantidade de condomínios Ativos.
	QtdCondomInativos      *int     `json:"QtdCondomInativos,omitempty"`      // Quantidade de condomínios inativos:
	QtdEconomAtivas        *int     `json:"QtdEconomAtivas,omitempty"`        // Quantidade de economias ativas.
	QtdEconomInativas      *int     `json:"QtdEconomInativas,omitempty"`      // Quantidade de economias inativas.
	EconomCaptadas         *int     `json:"EconomCaptadas,omitempty"`         // Quantidade de economias captadas.
	EconomRetiradas        *int     `json:"EconomRetiradas,omitempty"`        // Quantidade de economias retiradas.
	QtdBoletosEmitidos     *int     `json:"QtdBoletosEmitidos,omitempty"`     // Quantidade de boletos emitidos.
	VlrBoletosEmitidos     *types.Float64 `json:"VlrBoletosEmitidos,omitempty"`     // Valor total dos boletos.
	VlrTarifas             *types.Float64 `json:"VlrTarifas,omitempty"`             // Valor total da tarifa boleto.
	VlrTaxaCondom          *types.Float64 `json:"VlrTaxaCondom,omitempty"`          // Valor total da taxa de condomínio.
	VlrSegConteudoEmitido  *types.Float64 `json:"VlrSegConteudoEmitido,omitempty"`  // Valor total de seguro conteúdo emitido.
	VlrSegConteudoPago     *types.Float64 `json:"VlrSegConteudoPago,omitempty"`     // Valor total de seguro conteúdo pago.
	VlrSegConteudoRecebido *types.Float64 `json:"VlrSegConteudoRecebido,omitempty"` // Valor recebido de seguro conteúdo no mês.
	QtdBoletosNaoPagos     *int     `json:"QtdBoletosNaoPagos,omitempty"`     // Quantidade de boletos não pagos.
	VlrBoletosNaoPagos     *types.Float64 `json:"VlrBoletosNaoPagos,omitempty"`     // Valor total de boletos não pagos.
}

type RequestResponseBodyInfosExtras struct {
	VlrTaxaAReceberTotal          *types.Float64 `json:"VlrTaxaAReceberTotal,omitempty"`          // Valor total a receber de taxa de todas as economias ativas.
	QtdEconomAdimplentes          *int     `json:"QtdEconomAdimplentes,omitempty"`          // Quantidade total de economias adimplentes.
	VlrTaxaAReceberAdimplentes    *types.Float64 `json:"VlrTaxaAReceberAdimplentes,omitempty"`    // Valor total a receber em taxas de todas as economias adimplentes.
	QtdEconomInadimplentes        *int     `json:"QtdEconomInadimplentes,omitempty"`        // Quantidade de economias inadimplentes no momento.
	VlrEconomInadimplentes        *types.Float64 `json:"VlrEconomInadimplentes,omitempty"`        // Valor total a receber de economias inadimplentes.
	QtdEconomInadimpExtraJudicial *int     `json:"QtdEconomInadimpExtraJudicial,omitempty"` // Quantidade total de economias inadimplentes ? Ação Extra Judicial.
	VlrEconomInadimpExtraJudicial *types.Float64 `json:"VlrEconomInadimpExtraJudicial,omitempty"` // Valor total a receber das economias inadimplentes ? Ação Extra Judicial.
	QtdEconomInadimpJudicial      *int     `json:"QtdEconomInadimpJudicial,omitempty"`      // Quantidade de economias inadimplentes - Ação Judicial.
	VlrEconomInadimpJudicial      *types.Float64 `json:"VlrEconomInadimpJudicial,omitempty"`      // Valor total a receber das economias inadimplentes ? Ação Judicial.
}

type RequestResponseBodyTipoBoletos struct {
	Descricao   *string                                    `json:"Descricao,omitempty"`   // Tipos de boletos.
	Bancos      *[]RequestResponseBodyTipoBoletosBancos    `json:"Bancos,omitempty"`      //
	ResumoGeral *RequestResponseBodyTipoBoletosResumoGeral `json:"ResumoGeral,omitempty"` //
}

type RequestResponseBodyTipoBoletosBancos struct {
	CodBanco      *int                                           `json:"CodBanco,omitempty"`      // Código do banco.
	NomeBanco     *string                                        `json:"NomeBanco,omitempty"`     // Nome do banco.
	ContaCorrente *string                                        `json:"ContaCorrente,omitempty"` // Conta Corrente.
	Boletos       *[]RequestResponseBodyTipoBoletosBancosBoletos `json:"Boletos,omitempty"`       //
	Totais        *RequestResponseBodyTipoBoletosBancosTotais    `json:"Totais,omitempty"`        //
}

type RequestResponseBodyTipoBoletosBancosBoletos struct {
	Data      *string  `json:"Data,omitempty"`      // Data.
	QtdTotal  *int     `json:"QtdTotal,omitempty"`  // Quantidade de boletos emitidos no dia.
	VlrTotal  *types.Float64 `json:"VlrTotal,omitempty"`  // Valor total de boletos emitidos no dia.
	QtdNormal *int     `json:"QtdNormal,omitempty"` // Quantidade de boletos normais/extras no dia.
	VlrNormal *types.Float64 `json:"VlrNormal,omitempty"` // Valor dos boletos normais/extras no dia.
	QtdRetido *int     `json:"QtdRetido,omitempty"` // Quantidades de boletos rettidos no dia.
	VlrRetido *types.Float64 `json:"VlrRetido,omitempty"` // Valor dos boletos retidos no dia.
}

type RequestResponseBodyTipoBoletosBancosTotais struct {
	QtdTotal  *int     `json:"QtdTotal,omitempty"`  // Quantidade de boletos emitidos no dia.
	VlrTotal  *types.Float64 `json:"VlrTotal,omitempty"`  // Valor total de boletos emitidos no dia.
	QtdNormal *int     `json:"QtdNormal,omitempty"` // Quantidade de boletos normais/extras no dia.
	VlrNormal *types.Float64 `json:"VlrNormal,omitempty"` // Valor dos boletos normais/extras no dia.
	QtdRetido *int     `json:"QtdRetido,omitempty"` // Quantidades de boletos rettidos no dia.
	VlrRetido *types.Float64 `json:"VlrRetido,omitempty"` // Valor dos boletos retidos no dia.
}

type RequestResponseBodyTipoBoletosResumoGeral struct {
	Boletos *[]RequestResponseBodyTipoBoletosResumoGeralBoletos //
	Totais  *RequestResponseBodyTipoBoletosResumoGeralTotais    //
}

type RequestResponseBodyTipoBoletosResumoGeralBoletos struct {
	Data      *string  `json:"Data,omitempty"`      // Data.
	QtdTotal  *int     `json:"QtdTotal,omitempty"`  // Quantidade de boletos emitidos no dia.
	VlrTotal  *types.Float64 `json:"VlrTotal,omitempty"`  // Valor total de boletos emitidos no dia.
	QtdNormal *int     `json:"QtdNormal,omitempty"` // Quantidade de boletos normais/extras no dia.
	VlrNormal *types.Float64 `json:"VlrNormal,omitempty"` // Valor dos boletos normais/extras no dia.
	QtdRetido *int     `json:"QtdRetido,omitempty"` // Quantidades de boletos rettidos no dia.
	VlrRetido *types.Float64 `json:"VlrRetido,omitempty"` // Valor dos boletos retidos no dia.
}

type RequestResponseBodyTipoBoletosResumoGeralTotais struct {
	QtdTotal  *int     `json:"QtdTotal,omitempty"`  // Quantidade de boletos emitidos no dia.
	VlrTotal  *types.Float64 `json:"VlrTotal,omitempty"`  // Valor total de boletos emitidos no dia.
	QtdNormal *int     `json:"QtdNormal,omitempty"` // Quantidade de boletos normais/extras no dia.
	VlrNormal *types.Float64 `json:"VlrNormal,omitempty"` // Valor dos boletos normais/extras no dia.
	QtdRetido *int     `json:"QtdRetido,omitempty"` // Quantidades de boletos rettidos no dia.
	VlrRetido *types.Float64 `json:"VlrRetido,omitempty"` // Valor dos boletos retidos no dia.
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
