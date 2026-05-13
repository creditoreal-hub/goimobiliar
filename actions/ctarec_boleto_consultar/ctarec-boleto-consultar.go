package ctarec_boleto_consultar

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

var ACTION = "CTAREC_BOLETO_CONSULTAR"

type ActionInput struct {
	NossoNumero *string `json:"NossoNumero,omitempty"` // *Número de identificação bancário.
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
	OrigemCobranca          *string   `json:"OrigemCobranca,omitempty"`          //	String(1)	Locação/Condominio.
	DocCapaId               *int      `json:"DocCapaId,omitempty"`               //	Number(8)	Código do boleto no sistema.
	DOCRetido               *string   `json:"DOCRetido,omitempty"`               //	String(1)	Indica se é um boleto/DOC retido.
	Cancelado               *string   `json:"Cancelado,omitempty"`               //	String(1)	Indica se é um boleto/DOC cancelado.
	CodPessoa               *int      `json:"CodPessoa,omitempty"`               //	Number(7)	Código de pessoa do sacado.
	CobrRegAcordoVerParcAnt *string   `json:"CobrRegAcordoVerParcAnt,omitempty"` //	String(1)
	IdAcordo                *int      `json:"IdAcordo,omitempty"`                //	Number(8)	Código de identificação do acordo.
	DataVencAcordo          *string   `json:"DataVencAcordo,omitempty"`          //	Date	Data de vencimento do acordo.
	FilialNome              *string   `json:"FilialNome,omitempty"`              //	String(45)	Nome da filial.
	FilialEnd               *string   `json:"FilialEnd,omitempty"`               //	String(95)	Endereço da filial.
	FilialCnpj              *int      `json:"FilialCnpj,omitempty"`              //	Number(14)	Cnpj da filial.
	DataVenc                *string   `json:"DataVenc,omitempty"`                //	Date	Data de vencimento do boleto.
	DataPagamento           *string   `json:"DataPagamento,omitempty"`           //	Date	Data do pagamento.
	TipoDOC                 *string   `json:"TipoDOC,omitempty"`                 //	String(1)	Tipo de boleto/DOC.
	FilialCidade            *string   `json:"FilialCidade,omitempty"`            //	String(40)	Cidade da filial.
	IdCodBanco              *string   `json:"IdCodBanco,omitempty"`              //	String(5)	Código do banco com dígito verificador.
	LinhaDigitavel          *string   `json:"LinhaDigitavel,omitempty"`          //	String(60)	Linha digitável do boleto.
	PixQrCode               *string   `json:"PixQrCode,omitempty"`               //	String(390)	Qr Code do Pix vinculado ao boleto.
	VlrDocumento            *float64  `json:"VlrDocumento,omitempty"`            //	Number(12,2)	Valor do documento.
	NossoNumeroOrig         *string   `json:"NossoNumeroOrig,omitempty"`         //	String(13)	Nosso Numero original.
	LocalPagamento          *string   `json:"LocalPagamento,omitempty"`          //	String(80)	Local de pagamento.
	NomeCedente             *string   `json:"NomeCedente,omitempty"`             //	String(70)	Nome do cedente.
	CodCedente              *string   `json:"CodCedente,omitempty"`              //	String(15)	Código do cedente.
	DataDocumento           *string   `json:"DataDocumento,omitempty"`           //	Date	Data do documento.
	DataProcessamento       *string   `json:"DataProcessamento,omitempty"`       //	Date	Data de processamento.
	NumeroDOC               *string   `json:"NumeroDOC,omitempty"`               //	String	Número do documento.
	NossoNumero             *string   `json:"NossoNumero,omitempty"`             //	String(13)	Número de identificação bancário.
	Carteira                *string   `json:"Carteira,omitempty"`                //	String(7)	Carteira bancária.
	EspecieDOC              *string   `json:"EspecieDOC,omitempty"`              //	String(13)	Espécie do documento.
	Aceite                  *string   `json:"Aceite,omitempty"`                  //	String(13)	Aceite do documento.
	UsoBanco                *string   `json:"UsoBanco,omitempty"`                //	String(13)	Informações de uso do banco.
	Moeda                   *string   `json:"Moeda,omitempty"`                   //	String(20)	Moeda do documento.
	VlrAcrescOutr           *float64  `json:"VlrAcrescOutr,omitempty"`           //	Number(15,2)	Valor de outros acréscimos.
	VlrDesconto             *string   `json:"VlrDesconto,omitempty"`             //	String(15)	Valor de desconto.
	VlrDescOutr             *float64  `json:"VlrDescOutr,omitempty"`             //	Number(12,2)	Valor de outros descontos.
	VlrMulta                *string   `json:"VlrMulta,omitempty"`                //	String	Valor da multa mais juros.
	VlrSegCont              *float64  `json:"VlrSegCont,omitempty"`              //	Number(12,2)	Valor do seguro conteúdo.
	Sacado1                 *string   `json:"Sacado1,omitempty"`                 //	String	Primeira linha de informações do sacado.
	Sacado2                 *string   `json:"Sacado2,omitempty"`                 //	String	Segunda linha de informações do sacado.
	Sacado3                 *string   `json:"Sacado3,omitempty"`                 //	String	Terceira linha de informações do sacado.
	CodBarras               *string   `json:"CodBarras,omitempty"`               //	String(100)	Código de barras do boleto.
	Aviso                   *string   `json:"Aviso,omitempty"`                   //	String	Aviso do documento.
	DataLimitePagamento     *string   `json:"DataLimitePagamento,omitempty"`     //	Date	Data limite de pagamento do documento.
	DataTiraInadimplencia   *string   `json:"DataTiraInadimplencia,omitempty"`   //	Date	Data da retirada do boleto da inadimplencia.
	Instrucoes              *[]string `json:"Instrucoes,omitempty"`              //
	Detalhes                *[]string `json:"Detalhes,omitempty"`                //
	Informativos            *[]any    `json:"Informativos,omitempty"`            //
	Cabecalhos              *[]string `json:"Cabecalhos,omitempty"`              //
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
