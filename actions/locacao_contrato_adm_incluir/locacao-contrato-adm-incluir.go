package locacao_contrato_adm_incluir

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

var ACTION = "LOCACAO_CONTRATO_ADM_INCLUIR"

type ActionInput struct {
	DataVigInicial    *string                    `json:"DataVigInicial,omitempty"`    // *Data de vigência inicial deste contrato de administração.
	DataVigFinal      *string                    `json:"DataVigFinal,omitempty"`      // *Data de vigência final deste contrato de administração.
	Observacoes       *string                    `json:"Observacoes,omitempty"`       // Observações deste contrato de administração.
	CodFilial         *string                    `json:"CodFilial,omitempty"`         // Código da filial à qual o contrato pertence. Valor default é '001'.
	CodPessoaTitular  *int                       `json:"CodPessoaTitular,omitempty"`  // *Código da pessoa que é a titular deste contrato de administração.
	DataAssinatura    *string                    `json:"DataAssinatura,omitempty"`    // *Data de assinatura deste contrato de administração.
	Prazo             *int                       `json:"Prazo,omitempty"`             // *Prazo do contrato em meses.
	CodIntegracaoSist *string                    `json:"CodIntegracaoSist,omitempty"` // Código deste contrato de administração no sistema integrado/migrado.
	Proprietarios     *[]ActionInputProprietario `json:"Proprietarios,omitempty"`     // Lista de pessoas que são os proprietários do(s) imóvel(eis).
	Imoveis           *[]ActionInputImovel       `json:"Imoveis,omitempty"`           // Lista do(s) imóvel(eis) que pertencem a este contrato.
	Participacoes     *[]ActionInputParticipacao `json:"Participacoes,omitempty"`     // A lista de participações deve informar a divisão de renda entre os proprietários de cada imóvel do contrato. O somatório dos percentuais de participação de cada imóvel deve ser exatamente 100% e é sempre obrigatório mesmo que haja apenas um único proprietário com participação total.
}

type ActionInputProprietario struct {
	CodPessoa           *int    `json:"CodPessoa,omitempty"`           // *Código de pessoa do proprietário.
	IsentaTaxaDIMOB     *string `json:"IsentaTaxaDIMOB,omitempty"`     // *Indica se deve isentar da taxa de elaboração do DIMOB.
	IsentaJuros         *string `json:"IsentaJuros,omitempty"`         // *Indica se deve isentar dos juros.
	IsentaTarifaRemessa *string `json:"IsentaTarifaRemessa,omitempty"` // *Indica se deve isentar da tarifa para remessa.
	IsentaTaxaPorte     *string `json:"IsentaTaxaPorte,omitempty"`     // *Indica se deve isentar da taxa porte.
	CalcImpostos        *string `json:"CalcImpostos,omitempty"`        // Indica se deve calcular os impostos.
	EmiteNota           *string `json:"EmiteNota,omitempty"`           // Indica se deve emitir nota fiscal.
	CarneLeao           *string `json:"CarneLeao,omitempty"`           // *Indica se deve calcular valores relativos ao Carnê Leão.
	IsentaCOFINS        *string `json:"IsentaCOFINS,omitempty"`        // *Indica se deve isentar do COFINS.
	CalcIR              *string `json:"CalcIR,omitempty"`              // Indica se deve isentar do Imposto de Renda.
	SubstTributario     *string `json:"SubstTributario,omitempty"`     // Indica se é caso de substituição tributária.
	ExpInternet         *string `json:"ExpInternet,omitempty"`         // *Indica se deve exportar informações para o site na internet.
	ExpGrafica          *string `json:"ExpGrafica,omitempty"`          // *Indica se deve exportar demonstrativos para a gráfica.
	FormaCalc           *string `json:"FormaCalc,omitempty"`           // *Indica a periodicidade do calculo de juros.
	FormaPagto          *string `json:"FormaPagto,omitempty"`          // *Forma de pagamento ao proprietário.
	EmiteDemonstrativo  *string `json:"EmiteDemonstrativo,omitempty"`  // *Indica o tipo de demonstrativo para exibir ao proprietário.
	TipoEnderCorresp    *string `json:"TipoEnderCorresp,omitempty"`    // *Tipo do endereço da pessoa para enviar corresponência.
	CodPessoaProcurador *int    `json:"CodPessoaProcurador,omitempty"` // Código de pessoa do beneficiário.
	CodPessoaBenef      *int    `json:"CodPessoaBenef,omitempty"`      // *Código de pessoa do beneficiário.
	CodBanco            *int    `json:"CodBanco,omitempty"`            // Código do banco onde esta pessoa tem conta bancária.
	CodAgencia          *int    `json:"CodAgencia,omitempty"`          // Código da agencia onde esta pessoa tem conta bancária.
	ContaCorrente       *string `json:"ContaCorrente,omitempty"`       // Número da conta bancária desta pessoa.
	TipoConta           *string `json:"TipoConta,omitempty"`           // Tipo da conta bancária desta pessoa.
	IndiceMonetario     *string `json:"IndiceMonetario,omitempty"`     // Índice monetário para correção dos valores deste contrato.
	DimobInternet       *string `json:"DimobInternet,omitempty"`       // Indica se Demonstrativo de IR pode ser publicado no site.
}

type ActionInputImovel struct {
	CodImovel *int `json:"CodImovel,omitempty"` // *Código do imóvel.
}

type ActionInputParticipacao struct {
	CodImovel *int     `json:"CodImovel"` // *Código do imóvel.
	CodPessoa *int     `json:"CodPessoa"` // *Código de pessoa do proprietário.
	PercRenda *types.Float64 `json:"PercRenda"` // *Percentual de renda que esta pessoa possui neste imóvel (máximo de 100%).
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
	CodContratoAdm *int `json:"CodContratoAdm,omitempty"` // Código do contrato de administração.
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
