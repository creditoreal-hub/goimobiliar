package condom_economia_alterar

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

var ACTION = "CONDOM_ECONOMIA_ALTERAR"

type ActionInput struct {
	IdEconomia                     *int     `json:"IdEconomia,omitempty"`                     // *Chave principal da economia/unidade.
	CodEconomia                    *string  `json:"CodEconomia,omitempty"`                    // Código da economia/unidade no bloco.
	CodClasseImovel                *int     `json:"CodClasseImovel,omitempty"`                // Código da classe de imóvel.
	CodPessoaCondomino             *int     `json:"CodPessoaCondomino,omitempty"`             // Código de pessoa do condômino desta economia/unidade.
	CodPessoaLocat                 *int     `json:"CodPessoaLocat,omitempty"`                 // Código de pessoa do locatário desta economia/unidade.
	CodPessoaDebContaCondomino     *int     `json:"CodPessoaDebContaCondomino,omitempty"`     // Código de pessoa do condômino para débito em conta.
	CodPessoaDebContaLocat         *int     `json:"CodPessoaDebContaLocat,omitempty"`         // Código de pessoa do locatário para débito em conta.
	CodFornecedorAdministradoraLoc *int     `json:"CodFornecedorAdministradoraLoc,omitempty"` // Código de fornecedor da administradora da locação.
	CodImovelNaAdministradoraLoc   *int     `json:"CodImovelNaAdministradoraLoc,omitempty"`   // Código do imóvel na locação desta administradora.
	CodCompensacaoIntegrada        *string  `json:"CodCompensacaoIntegrada,omitempty"`        // Código do imóvel para compensação integrada com outra administradora da locação.
	CodFornecAdvogado              *int     `json:"CodFornecAdvogado,omitempty"`              // Código de fornecedor do advogado de cobrança dos boletos.
	TarifaBoleto                   *string  `json:"TarifaBoleto,omitempty"`                   // Indica se o boleto tem tarifa.
	ValorTarifaBoleto              *types.Float64 `json:"ValorTarifaBoleto,omitempty"`              // Valor fixado da tarifa.
	QtdeDormitorios                *int     `json:"QtdeDormitorios,omitempty"`                // Quantidade de dormitórios.
	Fracao                         *types.Float64 `json:"Fracao,omitempty"`                         // Fracao da economia/unidade.
	EmiteExtrato                   *string  `json:"EmiteExtrato,omitempty"`                   // Indica qual tipo de extrato.
	ExportaLocacao                 *string  `json:"ExportaLocacao,omitempty"`                 // Indica se exporta para locação.
	EmiteEtiqueta                  *string  `json:"EmiteEtiqueta,omitempty"`                  // Indica se emite etiqueta.
	RetemBoleto                    *string  `json:"RetemBoleto,omitempty"`                    // Indica se deve reter boleto.
	ExtratoNoSite                  *string  `json:"ExtratoNoSite,omitempty"`                  // Indica se deve mostrar extrato no site.
	EnviarEmailBoleto              *string  `json:"EnviarEmailBoleto,omitempty"`              // Indica se deve enviar boleto por e-mail.
	GerarReciboAluguel             *string  `json:"GerarReciboAluguel,omitempty"`             // Indica se deve gerar recibo de locação.
	IsentarTaxaPorte               *string  `json:"IsentarTaxaPorte,omitempty"`               // Indica se deve isentar taxa porte.
	AssociarAdvogado               *string  `json:"AssociarAdvogado,omitempty"`               // Indica se deve associar um advogado aos boletos.
	InibirMsgInadimplenciaBoleto   *string  `json:"InibirMsgInadimplenciaBoleto,omitempty"`   // Indica se deve inibir mensagem de inadimplência no boleto.
	InibirCartaInadimplencia       *string  `json:"InibirCartaInadimplencia,omitempty"`       // Indica se deve inibir impressão da carta de inadimplência.
	InibirEmailInadimplencia       *string  `json:"InibirEmailInadimplencia,omitempty"`       // Indica se deve inibir envio por email da carta de inadimplência.
	InibirExportacao               *string  `json:"InibirExportacao,omitempty"`               // Indica se deve gerar recibo de locação.
	BloqueioNegativa               *string  `json:"BloqueioNegativa,omitempty"`               // Indica se deve bloquear a negativa de débitos.
	ObservacaoEconomia             *string  `json:"ObservacaoEconomia,omitempty"`             // Observação sobre esta economia/unidade.
	ObservacaoBoleto               *string  `json:"ObservacaoBoleto,omitempty"`               // Texto para constar nas observações do boleto.
	LocalEnderCobr                 *string  `json:"LocalEnderCobr,omitempty"`                 // Local do endereço de cobrança.
	LocalEnderCorresp              *string  `json:"LocalEnderCorresp,omitempty"`              // Local do endereço de correpondência.
	Ativa                          *string  `json:"Ativa,omitempty"`                          // Indica se está ativa.
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
	IdEconomia *int `json:"IdEconomia,omitempty"` // Chave principal da economia/unidade.
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
