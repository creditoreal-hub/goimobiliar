package locacao_seguro_incluir

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

var ACTION = "LOCACAO_SEGURO_INCLUIR"

type ActionInput struct {
	CodImovel              *int                `json:"CodImovel,omitempty"`              // *Código do imóvel.
	VigenciaInicial        *string             `json:"VigenciaInicial,omitempty"`        // *Data inicial da vigência do seguro contratado.
	VigenciaFinal          *string             `json:"VigenciaFinal,omitempty"`          // *Data final da vigência do seguro contratado.
	Apolice                *string             `json:"Apolice,omitempty"`                // Número da apólice.
	CodRisco               *string             `json:"CodRisco,omitempty"`               //
	Proposta               *string             `json:"Proposta,omitempty"`               // Número da proposta.
	TipoSeguro             *string             `json:"TipoSeguro,omitempty"`             // *Tipo do seguro contratado (Fiança, Incêndio, etc.).
	CodSeguradora          *int                `json:"CodSeguradora,omitempty"`          // *Código interno da seguradora no Imibiliar.
	CodCorretor            *int                `json:"CodCorretor,omitempty"`            // Código interno da empresa corretora do seguro no Imobiliar.
	ValorSegurado          *types.Float64            `json:"ValorSegurado,omitempty"`          // *Valor segurado contratado.
	ValorPremioTotal       *types.Float64            `json:"ValorPremioTotal,omitempty"`       // *Valor total do prêmio a ser pago à seguradora.
	LancarContasPagar      *string             `json:"LancarContasPagar,omitempty"`      // *Este campo indica se deve fazer a incluisão das parcelas no Contas à Pagar.
	CompetenciaInicial     *string             `json:"CompetenciaInicial,omitempty"`     // Competência inicial da cobrança das parcelas.
	DataVencimentoInicial  *string             `json:"DataVencimentoInicial,omitempty"`  // Data de vencimento da primeira parcela.
	NumeroParcelaResidual  *int                `json:"NumeroParcelaResidual,omitempty"`  // Número da parcela a aplicar o valor residual (informar 1).
	ValorParcela           *types.Float64            `json:"ValorParcela,omitempty"`           // Valor da parcela.
	NumeroParcelas         *int                `json:"NumeroParcelas,omitempty"`         // Número de parcelas.
	ValorParcelasRestantes *types.Float64            `json:"ValorParcelasRestantes,omitempty"` //
	Lista                  *[]ActionInputLista `json:"Lista,omitempty"`                  // *Lista de parcelas de valores prédefinidos.
}

type ActionInputLista struct {
	NumeroParcela *int `json:"NumeroParcela,omitempty"` // *Número da parcela.
	ValorParcela  *int `json:"ValorParcela,omitempty"`  // *Valor da parcela.
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

type RequestResponseBody struct{}

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
