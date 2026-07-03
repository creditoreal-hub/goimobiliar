package ctarec_boleto_inadimplente_2via

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

var ACTION = "CTAREC_BOLETO_INADIMPLENTE_2VIA"

type ActionInput struct {
	NossoNumero          *string        `json:"NossoNumero,omitempty"`          // *Número de identificação bancário.
	DataLimitePagamento  *string        `json:"DataLimitePagamento,omitempty"`  // *Data limite de pagamento do documento.
	NroDiasIniVencto     *types.Float64 `json:"NroDiasIniVencto,omitempty"`     //
	NroDiasFimVencto     *types.Float64 `json:"NroDiasFimVencto,omitempty"`     //
	Email                *string        `json:"Email,omitempty"`                // E-mail da pessoa.
	InibirCobrRegistrada *string        `json:"InibirCobrRegistrada,omitempty"` // Valor default é 'N'.
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
	DocCapaId               *int                              `json:"DocCapaId,omitempty"`               // Código do boleto no sistema.
	DOCRetido               *string                           `json:"DOCRetido,omitempty"`               // Indica se é um boleto/DOC retido.
	CodPessoa               *int                              `json:"CodPessoa,omitempty"`               // Código de pessoa do sacado.
	CobrRegAcordoVerParcAnt *string                           `json:"CobrRegAcordoVerParcAnt,omitempty"` //
	IdAcordo                *int                              `json:"IdAcordo,omitempty"`                // Código de identificação do acordo.
	DataVencAcordo          *string                           `json:"DataVencAcordo,omitempty"`          // Data de vencimento do acordo.
	FilialNome              *string                           `json:"FilialNome,omitempty"`              // Nome da filial.
	FilialEnd               *string                           `json:"FilialEnd,omitempty"`               // Endereço da filial.
	FilialCnpj              *types.CpfCnpj                    `json:"FilialCnpj,omitempty"`              // Cnpj da filial.
	DataVenc                *string                           `json:"DataVenc,omitempty"`                // Data de vencimento do boleto.
	FilialCidade            *string                           `json:"FilialCidade,omitempty"`            // Cidade da filial.
	IdCodBanco              *string                           `json:"IdCodBanco,omitempty"`              // Código do banco com dígito verificador.
	LinhaDigitavel          *string                           `json:"LinhaDigitavel,omitempty"`          // Linha digitável do boleto.
	VlrDocumento            *types.Float64                    `json:"VlrDocumento,omitempty"`            // Valor do documento.
	NossoNumeroOrig         *string                           `json:"NossoNumeroOrig,omitempty"`         // Nosso Numero original.
	LocalPagamento          *string                           `json:"LocalPagamento,omitempty"`          // Local de pagamento.
	NomeCedente             *string                           `json:"NomeCedente,omitempty"`             // Nome do cedente.
	CodCedente              *string                           `json:"CodCedente,omitempty"`              // Código do cedente.
	DataDocumento           *string                           `json:"DataDocumento,omitempty"`           // Data do documento.
	DataProcessamento       *string                           `json:"DataProcessamento,omitempty"`       // Data de processamento.
	NumeroDOC               *string                           `json:"NumeroDOC,omitempty"`               // Número do documento.
	NossoNumero             *string                           `json:"NossoNumero,omitempty"`             // Número de identificação bancário.
	Carteira                *string                           `json:"Carteira,omitempty"`                // Carteira bancária.
	EspecieDOC              *string                           `json:"EspecieDOC,omitempty"`              // Espécie do documento.
	Aceite                  *string                           `json:"Aceite,omitempty"`                  // Aceite do documento.
	UsoBanco                *string                           `json:"UsoBanco,omitempty"`                // Informações de uso do banco.
	Moeda                   *string                           `json:"Moeda,omitempty"`                   // Moeda do documento.
	VlrAcrescOutr           *types.Float64                    `json:"VlrAcrescOutr,omitempty"`           // Valor de outros acréscimos.
	VlrDesconto             *string                           `json:"VlrDesconto,omitempty"`             // Valor de desconto.
	VlrDescOutr             *types.Float64                    `json:"VlrDescOutr,omitempty"`             // Valor de outros descontos.
	VlrMulta                *string                           `json:"VlrMulta,omitempty"`                // Valor da multa mais juros.
	VlrSegCont              *types.Float64                    `json:"VlrSegCont,omitempty"`              // Valor do seguro conteúdo.
	Sacado1                 *string                           `json:"Sacado1,omitempty"`                 // Primeira linha de informações do sacado.
	Sacado2                 *string                           `json:"Sacado2,omitempty"`                 // Segunda linha de informações do sacado.
	Sacado3                 *string                           `json:"Sacado3,omitempty"`                 // Terceira linha de informações do sacado.
	CodBarras               *string                           `json:"CodBarras,omitempty"`               // Código de barras do boleto.
	Aviso                   *string                           `json:"Aviso,omitempty"`                   // Aviso do documento.
	DataLimitePagamento     *string                           `json:"DataLimitePagamento,omitempty"`     // Data limite de pagamento do documento.
	Instrucoes              *[]RequestResponseBodyInstrucao   `json:"Instrucoes,omitempty"`              //
	Detalhes                *[]RequestResponseBodyDetalhe     `json:"Detalhes,omitempty"`                //
	Informativos            *[]RequestResponseBodyInformativo `json:"Informativos,omitempty"`            //
	Cabecalhos              *[]RequestResponseBodyCabecalho   `json:"Cabecalhos,omitempty"`              //
}

type RequestResponseBodyInstrucao struct {
	Instrucao *string `json:"Instrucao,omitempty"` // Linha de instrução.
}
type RequestResponseBodyDetalhe struct {
	Detalhe *string `json:"Detalhe,omitempty"` // Linha de detalhe.
}
type RequestResponseBodyInformativo struct {
	Informativo *string `json:"Informativo,omitempty"` // Linha de informação.
}
type RequestResponseBodyCabecalho struct {
	Cabecalho *string `json:"Cabecalho,omitempty"` // Linha de cabeçalho.
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
