package comerc_interessado_consultar

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

var ACTION = "COMERC_INTERESSADO_CONSULTAR"

type ActionInput struct {
	CodInteressado *int `json:"CodInteressado,omitempty"` // *Código do Interessado.
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
	CodInteressado      *int           `json:"CodInteressado,omitempty"`      // Código do Interessado.
	Nome                *string        `json:"Nome,omitempty"`                // Nome do Interessado.
	TipoPessoa          *string        `json:"TipoPessoa,omitempty"`          // Tipo da pessoa.
	CpfCnpj             *types.CpfCnpj `json:"CpfCnpj,omitempty"`             // Se for tipo de pessoa física o valor é um CPF. Se for tipo de pessoa jurídica o valor é um CNPJ. Se o tipo de pessoa não for informado então este campo é vazio.
	RG                  *string        `json:"RG,omitempty"`                  // Número do documento de identificação da pessoa física. Não preencher se for pessoa jurídica.
	Ativo               *string        `json:"Ativo,omitempty"`               // Indica se está ativo.
	OrgaoExpedidor      *string        `json:"OrgaoExpedidor,omitempty"`      // Órgão que expediu o documento de identificação informado.
	DataNascimento      *string        `json:"DataNascimento,omitempty"`      // Data de nascimento da pessoa física ou de criação da pessoa jurídica.
	Celular             *string        `json:"Celular,omitempty"`             // Número de celular.
	Email               *string        `json:"Email,omitempty"`               // E-mail do interessado.
	Contato             *string        `json:"Contato,omitempty"`             // Informações de pessoa de contato.
	Observacao          *string        `json:"Observacao,omitempty"`          // Mensagem de Observação.
	DataCadastro        *string        `json:"DataCadastro,omitempty"`        // Data do cadastro no sistema.
	TipoEnder           *string        `json:"TipoEnder,omitempty"`           // Tipo de endereço.
	FormaEndereco       *int           `json:"FormaEndereco,omitempty"`       //
	CEP                 *int           `json:"CEP,omitempty"`                 // Número do CEP.
	TipoLograd          *string        `json:"TipoLograd,omitempty"`          // Tipo de logradouro abreviado ou por extenso ('R' ou 'RUA', 'AV' ou 'AVENIDA', etc.).
	Logradouro          *string        `json:"Logradouro,omitempty"`          // Logradouro do endereço. Deve ser informado apenas o nome sem o tipo de logradouro.
	Numero              *int           `json:"Numero,omitempty"`              // Número do endereço.
	Complemento         *string        `json:"Complemento,omitempty"`         // Complemento do endereço.
	Bairro              *string        `json:"Bairro,omitempty"`              // Bairro do endereço.
	Cidade              *string        `json:"Cidade,omitempty"`              // Cidade do endereço.
	UF                  *string        `json:"UF,omitempty"`                  // Sigla da Unidade Federativa do endereço.
	Telefone1           *string        `json:"Telefone1,omitempty"`           // Número de telefone principal.
	Ramal1              *string        `json:"Ramal1,omitempty"`              // Ramal do telefone principal.
	Telefone2           *string        `json:"Telefone2,omitempty"`           // Número de telefone alternativo.
	Ramal2              *string        `json:"Ramal2,omitempty"`              // Ramal do telefone alternativo.
	UsuarioId           *string        `json:"UsuarioId,omitempty"`           // Identificação do usuário.
	IdAgencia           *int           `json:"IdAgencia,omitempty"`           // Identificação da Agência de Cadastro.
	CodCadPessoa        *int           `json:"CodCadPessoa,omitempty"`        // Código do cadastro de pessoas (Quando Cadastrado).
	TipoDivulgacao      *string        `json:"TipoDivulgacao,omitempty"`      //	Tipo de divulgação que a pessoa chegou até a empresa.
	TipoComercializacao *string        `json:"TipoComercializacao,omitempty"` // Informa se a comercialização é Locação ou Venda.
	CodVeiculo          *string        `json:"CodVeiculo,omitempty"`          // Código veículo de comunicação.
	CodCorretor         *int           `json:"CodCorretor,omitempty"`         // Código do Corretor.
	NomeCorretor        *string        `json:"NomeCorretor,omitempty"`        // Nome do Corretor.
	ProcuraAtiva        *string        `json:"ProcuraAtiva,omitempty"`        //  se a pessoa está com procura de imóveis ativa.
	QualificaPessoa     *string        `json:"QualificaPessoa,omitempty"`     // Qualificação da Pessoa.
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
