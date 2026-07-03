package condom_lista_economias

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

var ACTION = "CONDOM_LISTA_ECONOMIAS"

type ActionInput struct {
	CodCondominio        *int    `json:"CodCondominio,omitempty"`        // *Código do condomínio.
	CodBloco             *string `json:"CodBloco,omitempty"`             // Código do bloco do condomínio.
	DataAlteracaoInicial *string `json:"DataAlteracaoInicial,omitempty"` // Seleção por data de alteração.
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
	CodCondominio    *int                        `json:"CodCondominio,omitempty"`    // Código do condomínio.
	NomeCondominio   *string                     `json:"NomeCondominio,omitempty"`   // Nome do condomínio.
	CodBlocoBase     *string                     `json:"CodBlocoBase,omitempty"`     // Bloco base/principal do condomínio.
	CodFilial        *int                        `json:"CodFilial,omitempty"`        // Código da filial.
	DiaVencimentoDoc *int                        `json:"DiaVencimentoDoc,omitempty"` // Dia de vencimento do boleto de condomínio.
	Assessor         *string                     `json:"Assessor,omitempty"`         // Código de usuário do assessor do condomínio.
	AssessorEmail    *string                     `json:"AssessorEmail,omitempty"`    // Email do assessor do condomínio.
	AssessorAgencia  *string                     `json:"AssessorAgencia,omitempty"`  // Agência do assessor do condomínio.
	TotaldeEconomias *int                        `json:"TotaldeEconomias,omitempty"` // Total de economias do condomínio.
	TotaldeBlocos    *int                        `json:"TotaldeBlocos,omitempty"`    // Total de blocos do condomínio.
	Blocos           *[]RequestResponseBodyBloco `json:"Blocos,omitempty"`           //
}

type RequestResponseBodyBloco struct {
	CodBloco      *string                             `json:"CodBloco"`            // Código do bloco do condomínio.
	NomeBloco     *string                             `json:"NomeBloco"`           // Nome de bloco/conta.
	QtdeEconomias *int                                `json:"QtdeEconomias"`       // Total de economias do bloco.
	Endereco      *string                             `json:"Endereco"`            // Endereço do condomínio.
	Bairro        *string                             `json:"Bairro"`              // Bairro do endereço.
	CEP           *int                                `json:"CEP"`                 // Número do CEP.
	NomeSindico   *string                             `json:"NomeSindico"`         // Nome do síndico.
	EmailSindico  *string                             `json:"EmailSindico"`        // E-mail do síndico.
	CPFSindico    *string                             `json:"CPFSindico"`          // CPF do síndico.
	ValorGas      *types.Float64                      `json:"ValorGas"`            // Valor de consumo de gas.
	ValorAgua     *types.Float64                      `json:"ValorAgua"`           // Valor de consumo de água.
	Economias     *[]RequestResponseBodyBlocoEconomia `json:"Economias,omitempty"` //
	Conselho      *[]RequestResponseBodyBlocoConselho `json:"Conselho,omitempty"`  //
}

type RequestResponseBodyBlocoEconomia struct {
	IdEconomia         *int                                        `json:"IdEconomia,omitempty"`         // Chave principal da economia/unidade.
	CodEconomia        *string                                     `json:"CodEconomia,omitempty"`        // Código da economia/unidade no bloco.
	CodPessoaCondomino *int                                        `json:"CodPessoaCondomino,omitempty"` // Código de pessoa do condômino desta economia/unidade.
	Nome               *string                                     `json:"Nome,omitempty"`               // Nome do condômino.
	Celular            *string                                     `json:"Celular,omitempty"`            // Número de celular do condomino.
	Fracao             *types.Float64                              `json:"Fracao,omitempty"`             // Fracao da economia/unidade.
	Email              *string                                     `json:"Email,omitempty"`              // E-mail do condômino.
	Locatario          *string                                     `json:"Locatario,omitempty"`          // Nome do locatário.
	Contato            *string                                     `json:"Contato,omitempty"`            // Informações de contato.
	CpfCnpj            *types.CpfCnpj                              `json:"CpfCnpj,omitempty"`            // CPF do condômino.
	Enderecos          *[]RequestResponseBodyBlocoEconomiaEndereco `json:"Enderecos,omitempty"`          //
}

type RequestResponseBodyBlocoEconomiaEndereco struct {
	TipoEndereco *string `json:"TipoEndereco,omitempty"` // Tipo do enderereco do condômino.
	Enderereco   *string `json:"Enderereco,omitempty"`   // Enderereco do condômino.
	Cidade       *string `json:"Cidade,omitempty"`       // Cidade do endereço.
	Bairro       *string `json:"Bairro,omitempty"`       // Bairro do endereço.
	CEP          *int    `json:"CEP,omitempty"`          // Número do CEP.
	UF           *string `json:"UF,omitempty"`           // Sigla da Unidade Federativa do endereço.
	Telefone1    *string `json:"Telefone1,omitempty"`    // Número de telefone principal.
	Telefone2    *string `json:"Telefone2,omitempty"`    // Número de telefone alternativo.
}

type RequestResponseBodyBlocoConselho struct {
	CodPessoa           *int    `json:"CodPessoa,omitempty"`           // Código da pessoa.
	Cargo               *string `json:"Cargo,omitempty"`               // Cargo no conselho de condomínio.
	InicioMandato       *string `json:"InicioMandato,omitempty"`       // Data do início do mandato.
	FinalMandato        *string `json:"FinalMandato,omitempty"`        // Data do final de mandato.
	SindicoProfissional *string `json:"SindicoProfissional,omitempty"` // Indicação de síndico profissional.
	CodFornecedor       *int    `json:"CodFornecedor,omitempty"`       // Código de fornecedor (se for o caso).
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
