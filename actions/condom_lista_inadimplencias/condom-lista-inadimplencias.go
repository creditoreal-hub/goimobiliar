package condom_lista_inadimplencias

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

var ACTION = "CONDOM_LISTA_INADIMPLENCIAS"

type ActionInput struct {
	CodCondominio                  *int    `json:"CodCondominio,omitempty"`                  // *Código do condomínio.
	CodBloco                       *string `json:"CodBloco,omitempty"`                       // Se informado o código do bloco então busca apenas a inadimplencia desse bloco senão busca toda a inadimplencia do condominio.
	IdEconomia                     *int    `json:"IdEconomia,omitempty"`                     // Se informada a chave da economia/unidade então busca apenas a inadimplencia dela senão busca toda a inadimplencia do condominio.
	IncluirDocsAcordo              *string `json:"IncluirDocsAcordo,omitempty"`              // Indica se deve incluir acordos. Valor default é 'N'.
	IncluirObsInadimplencia        *string `json:"IncluirObsInadimplencia,omitempty"`        // Indica se deve incluir observações do jurídico nos boletos inadimplentes. Valor default é 'N'.
	IncluirGarantidosInadimplencia *string `json:"IncluirGarantidosInadimplencia,omitempty"` // Indica se deve incluir boletos garantidos inadimplentes.
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
	Inadimplentes *[]RequestResponseBodyInadimplente `json:"Inadimplentes,omitempty"` //
}

type RequestResponseBodyInadimplente struct {
	DataVencimento       *string                                              `json:"DataVencimento,omitempty"`    //	Date	Data de vencimento do documento.
	CodBloco             *string                                              `json:"CodBloco,omitempty"`          //	String(3)	Se informado o código do bloco então busca apenas a inadimplencia desse bloco senão busca toda a inadimplencia do condominio.
	Economia             *string                                              `json:"Economia,omitempty"`          //	String	Identificação da economia.
	DescrClasseImovel    *string                                              `json:"DescrClasseImovel,omitempty"` //	String(50)	Descrição da classe de imóvel da economia/unidade.
	IdEconomia           *int                                                 `json:"IdEconomia,omitempty"`        //	Number(8)	Se informada a chave da economia/unidade então busca apenas a inadimplencia dela senão busca toda a inadimplencia do condominio.
	CodPessoa            *int                                                 `json:"CodPessoa,omitempty"`         //	Number(7)	Código da pessoa.
	Nome                 *string                                              `json:"Nome,omitempty"`              //	String(100)	Nome da pessoa.
	NossoNumero          *string                                              `json:"NossoNumero,omitempty"`       //	String(13)	Número de identificação bancário.
	TipoDOC              *string                                              `json:"TipoDOC,omitempty"`           //	String(1)	Tipo de boleto/DOC.
	Competencia          *string                                              `json:"Competencia,omitempty"`       //	String(7)	Competência do documento sem quitação.
	VlrDocumento         *float64                                             `json:"VlrDocumento,omitempty"`      //	Number(12,2)	Valor do documento.
	VlrCorrigido         *float64                                             `json:"VlrCorrigido,omitempty"`      //	Number(12,2)	Valor corrigido.
	Multa                *float64                                             `json:"Multa,omitempty"`             //	Number(12,2)	Multa sobre valor original.
	Juros                *float64                                             `json:"Juros,omitempty"`             //	Number(12,2)	Juros sobre valor original.
	Correcao             *float64                                             `json:"Correcao,omitempty"`          //	Number(12,2)	Correção monetária sobre valor original.
	VlrHonorarios        *float64                                             `json:"VlrHonorarios,omitempty"`     //	Number(12,2)	Valor dos honorários jurídicos.
	VlrCustas            *float64                                             `json:"VlrCustas,omitempty"`         //	Number(12,2)	Valor das custas jurídicas.
	VlrTotal             *float64                                             `json:"VlrTotal,omitempty"`          //	Number(12,2)	Valor total com honorários e custas.
	ObsJurNomeAdv        *string                                              `json:"ObsJur_NomeAdv,omitempty"`    //	String	Nome do advogado responsável pelas observações jurídicas.
	ObservacoesJuridicas *[]RequestResponseBodyInadimplenteObservacaoJuridica //
}

type RequestResponseBodyInadimplenteObservacaoJuridica struct {
	Observacao *string `json:"Observacao,omitempty"` // Observação do jurídico.
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
