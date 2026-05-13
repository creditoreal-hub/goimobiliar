package condom_economia_consultar

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

var ACTION = "CONDOM_ECONOMIA_CONSULTAR"

type ActionInput struct {
	IdEconomia *int `json:"IdEconomia,omitempty"` // *Chave principal da economia/unidade.
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
	IdEconomia                     *int     `json:"IdEconomia,omitempty"`                     // Chave principal da economia/unidade.
	CodCondominio                  *int     `json:"CodCondominio,omitempty"`                  // Código do condomínio.
	CodBloco                       *string  `json:"CodBloco,omitempty"`                       // Código do bloco do condomínio.
	CodEconomia                    *string  `json:"CodEconomia,omitempty"`                    // Código da economia/unidade no bloco.
	CodClasseImovel                *int     `json:"CodClasseImovel,omitempty"`                // Código da classe de imóvel.
	DescrClasseImovel              *string  `json:"DescrClasseImovel,omitempty"`              // Descrição da classe de imóvel da economia/unidade.
	CodPessoaCondomino             *int     `json:"CodPessoaCondomino,omitempty"`             // Código de pessoa do condômino desta economia/unidade.
	CodPessoaDebContaCondomino     *int     `json:"CodPessoaDebContaCondomino,omitempty"`     // Código de pessoa do condômino para débito em conta.
	CodPessoaDebContaLocat         *int     `json:"CodPessoaDebContaLocat,omitempty"`         // Código de pessoa do locatário para débito em conta.
	Nome                           *string  `json:"Nome,omitempty"`                           // Nome do condômino.
	Celular                        *string  `json:"Celular,omitempty"`                        // Número de celular do condomino.
	Email                          *string  `json:"Email,omitempty"`                          // E-mail do condômino.
	CodPessoaLocat                 *int     `json:"CodPessoaLocat,omitempty"`                 // Código de pessoa do locatário desta economia/unidade.
	NomeLocat                      *string  `json:"NomeLocat,omitempty"`                      // Nome do locatário.
	Contato                        *string  `json:"Contato,omitempty"`                        // Informações de contato.
	TipoPessoa                     *string  `json:"TipoPessoa,omitempty"`                     // Tipo da pessoa.
	CpfCnpj                        *string  `json:"CpfCnpj,omitempty"`                        // CPF/CNPJ do condômino.
	QtdeDormitorios                *int     `json:"QtdeDormitorios,omitempty"`                // Quantidade de dormitórios.
	Fracao                         *types.Float64 `json:"Fracao,omitempty"`                         // Fracao da economia/unidade.
	EmiteExtrato                   *string  `json:"EmiteExtrato,omitempty"`                   // Indica qual tipo de extrato.
	ExportaLocacao                 *string  `json:"ExportaLocacao,omitempty"`                 // Indica se exporta para locação.
	EmiteEtiqueta                  *string  `json:"EmiteEtiqueta,omitempty"`                  // Indica se emite etiqueta.
	TarifaBoleto                   *string  `json:"TarifaBoleto,omitempty"`                   // Indica se o boleto tem tarifa.
	ValorTarifaBoleto              *types.Float64 `json:"ValorTarifaBoleto,omitempty"`              // Valor fixado da tarifa.
	CodFornecedorAdministradoraLoc *int     `json:"CodFornecedorAdministradoraLoc,omitempty"` // Código de fornecedor da administradora da locação.
	CodImovelNaAdministradoraLoc   *int     `json:"CodImovelNaAdministradoraLoc,omitempty"`   // Código do imóvel na locação desta administradora.
	CodCompensacaoIntegrada        *string  `json:"CodCompensacaoIntegrada,omitempty"`        // Código do imóvel para compensação integrada com outra administradora da locação.
	RetemBoleto                    *string  `json:"RetemBoleto,omitempty"`                    // Indica se deve reter boleto.
	ExtratoNoSite                  *string  `json:"ExtratoNoSite,omitempty"`                  // Indica se deve mostrar extrato no site.
	EnviarEmailBoleto              *string  `json:"EnviarEmailBoleto,omitempty"`              // Indica se deve enviar boleto por e-mail.
	GerarReciboAluguel             *string  `json:"GerarReciboAluguel,omitempty"`             // Indica se deve gerar recibo de locação.
	IsentarTaxaPorte               *string  `json:"IsentarTaxaPorte,omitempty"`               // Indica se deve isentar taxa porte.
	AssociarAdvogado               *string  `json:"AssociarAdvogado,omitempty"`               // Indica se deve associar um advogado aos boletos.
	CodFornecAdvogado              *int     `json:"CodFornecAdvogado,omitempty"`              // Código de fornecedor do advogado de cobrança dos boletos.
	InibirMsgInadimplenciaBoleto   *string  `json:"InibirMsgInadimplenciaBoleto,omitempty"`   // Indica se deve inibir mensagem de inadimplência no boleto.
	InibirCartaInadimplencia       *string  `json:"InibirCartaInadimplencia,omitempty"`       // Indica se deve inibir impressão da carta de inadimplência.
	InibirEmailInadimplencia       *string  `json:"InibirEmailInadimplencia,omitempty"`       // Indica se deve inibir envio por email da carta de inadimplência.
	InibirExportacao               *string  `json:"InibirExportacao,omitempty"`               // Indica se deve gerar recibo de locação.
	ObservacaoEconomia             *string  `json:"ObservacaoEconomia,omitempty"`             // Observação sobre esta economia/unidade.
	ObservacaoBoleto               *string  `json:"ObservacaoBoleto,omitempty"`               // Texto para constar nas observações do boleto.
	LocalEnderCobr                 *string  `json:"LocalEnderCobr,omitempty"`                 // Local do endereço de cobrança.
	TipoLogradCobr                 *string  `json:"TipoLogradCobr,omitempty"`                 // Tipo de logradouro abreviado ou por extenso ('R' ou 'RUA', 'AV' ou 'AVENIDA', etc.).
	LogradouroCobr                 *string  `json:"LogradouroCobr,omitempty"`                 // Logradouro do endereço de cobrança.
	NumeroCobr                     *int     `json:"NumeroCobr,omitempty"`                     // Número do endereço.
	ComplementoCobr                *string  `json:"ComplementoCobr,omitempty"`                // Complemento do endereço.
	CidadeCobr                     *string  `json:"CidadeCobr,omitempty"`                     // Cidade do endereço.
	BairroCobr                     *string  `json:"BairroCobr,omitempty"`                     // Bairro do endereço.
	CEPCobr                        *int     `json:"CEPCobr,omitempty"`                        // Número do CEP.
	UFCobr                         *string  `json:"UFCobr,omitempty"`                         // Sigla da Unidade Federativa do endereço.
	LocalEnderCorresp              *string  `json:"LocalEnderCorresp,omitempty"`              // Local do endereço de correpondência.
	TipoLogradCorresp              *string  `json:"TipoLogradCorresp,omitempty"`              // Tipo de logradouro abreviado ou por extenso ('R' ou 'RUA', 'AV' ou 'AVENIDA', etc.).
	LogradouroCorresp              *string  `json:"LogradouroCorresp,omitempty"`              // Logradouro do endereço de correpondência.
	NumeroCorresp                  *int     `json:"NumeroCorresp,omitempty"`                  // Número do endereço.
	ComplementoCorresp             *string  `json:"ComplementoCorresp,omitempty"`             // Complemento do endereço.
	CidadeCorresp                  *string  `json:"CidadeCorresp,omitempty"`                  // Cidade do endereço.
	BairroCorresp                  *string  `json:"BairroCorresp,omitempty"`                  // Bairro do endereço.
	CEPCorresp                     *int     `json:"CEPCorresp,omitempty"`                     // Número do CEP.
	UFCorresp                      *string  `json:"UFCorresp,omitempty"`                      // Sigla da Unidade Federativa do endereço.
	Ativa                          *string  `json:"Ativa,omitempty"`                          // Indica se está ativa.
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
