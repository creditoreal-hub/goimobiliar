package ctarec_boleto_acordo_calcular

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

var ACTION = "CTAREC_BOLETO_ACORDO_CALCULAR"

type ActionInput struct {
	DocCapaIds           *string               `json:"DocCapaIds,omitempty"`           // *Lista de códigos de boletos (DocCapaId) que devem entrar no acordo separados por virgula (,).
	DataVencPrimeiraParc *string               `json:"DataVencPrimeiraParc,omitempty"` // *Data de vencimento da primeira parcela.
	DataVencSegundaParc  *string               `json:"DataVencSegundaParc,omitempty"`  // Data de vencimento da segunda parcela.
	QtdParcelas          *float64              `json:"QtdParcelas,omitempty"`          // *Quantidade de parcelas do acordo.
	FormaLancto          *string               `json:"FormaLancto,omitempty"`          // *Forma de lançamento no sistema.
	FormaCobranca        *string               `json:"FormaCobranca,omitempty"`        // *Forma de cobrança.
	TipoCorrecao         *string               `json:"TipoCorrecao,omitempty"`         // *
	TipoAcordo           *string               `json:"TipoAcordo,omitempty"`           // *Código de identificação do acordo.
	Complemento          *string               `json:"Complemento,omitempty"`          // *Texto que identifica os boletos originais do acordo. Ex.: "Venctos 10/05/20yy a 10/08/20yy.".
	VlrCustas            *float64              `json:"VlrCustas,omitempty"`            //
	VlrHonorarios        *float64              `json:"VlrHonorarios,omitempty"`        //
	PercHonorarios       *float64              `json:"PercHonorarios,omitempty"`       // Percentual de honorários (a ser dividido entre as parcelas do acordo).
	VlrMulta             *float64              `json:"VlrMulta,omitempty"`             //
	VlrMultaProp         *float64              `json:"VlrMultaProp,omitempty"`         //
	VlrJuros             *float64              `json:"VlrJuros,omitempty"`             // Valor total de juros (a ser dividido entre as parcelas do acordo). Se não for informado, o sistema irá apurar conforme tempo de atraso dos boletos originais.
	VlrCorrecao          *float64              `json:"VlrCorrecao,omitempty"`          // Valor total de correção (a ser dividido entre as parcelas do acordo). Se não for informado, o sistema irá apurar conforme tempo de atraso dos boletos originais.
	PercJuros            *float64              `json:"PercJuros,omitempty"`            // Percentual de juros se atraso de boleto. Se não for informado, o sistema assumirá a cobrança tradicional de juros do condomínio.
	PercMulta            *float64              `json:"PercMulta,omitempty"`            // Percentual de multa se atraso de boleto. Se não for informado, o sistema assumirá a cobrança tradicional de multa do condomínio.
	HonorariosCC         *string               `json:"HonorariosCC,omitempty"`         // Indica se deverá ou não lançar honorários em conta corrente na quitação da parcela do acordo. Valor default é 'N'),TransactField(Honorarios_CC.
	ExportarDoc          *string               `json:"ExportarDoc,omitempty"`          // *Indica se deverá ou não efetuar exportação dos boletos do acordo.
	Parcelas             *[]ActionInputParcela `json:"Parcelas,omitempty"`             //
}

type ActionInputParcela struct {
	Valor         *float64 `json:"Valor,omitempty"`         // *Valor de cada parcela.
	VlrHonorarios *float64 `json:"VlrHonorarios,omitempty"` //
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
	Boletos *[]RequestResponseBodyBoleto `json:"Boletos,omitempty"` //
}

type RequestResponseBodyBoleto struct {
	DataVenc      *string  `json:"DataVenc,omitempty"`      // Data de vencimento do boleto.
	TipoDOC       *string  `json:"TipoDOC,omitempty"`       // Tipo de boleto/DOC.
	Complemento   *string  `json:"Complemento,omitempty"`   // Texto que identifica os boletos originais do acordo. Ex.: "Venctos 10/05/20yy a 10/08/20yy.".
	Valor         *float64 `json:"Valor,omitempty"`         // Valor de cada parcela.
	VlrHonorarios *float64 `json:"VlrHonorarios,omitempty"` //
	VlrCustas     *float64 `json:"VlrCustas,omitempty"`     //
	VlrMulta      *float64 `json:"VlrMulta,omitempty"`      //
	VlrMultaProp  *float64 `json:"VlrMultaProp,omitempty"`  //
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
