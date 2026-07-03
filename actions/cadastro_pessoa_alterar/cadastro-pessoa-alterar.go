package cadastro_pessoa_alterar

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

var ACTION = "CADASTRO_PESSOA_ALTERAR"

type ActionInput struct {
	CodPessoa         *int                   `json:"CodPessoa,omitempty"`         // *Código da pessoa.
	Nome              *string                `json:"Nome,omitempty"`              // Nome da pessoa.
	NomePai           *string                `json:"NomePai,omitempty"`           // Nome do pai da pessoa física.
	NomeMae           *string                `json:"NomeMae,omitempty"`           // Nome da mãe da pessoa física.
	PIS               *string                `json:"PIS,omitempty"`               // PIS da pessoa da pessoa física.
	Nacionalidade     *string                `json:"Nacionalidade,omitempty"`     // Nacionalidade da pessoa no padrão do e-Social.
	CodNacionalidade  *int                   `json:"CodNacionalidade,omitempty"`  // Código de nacionalidade da pessoa no e-Social.
	Naturalidade      *string                `json:"Naturalidade,omitempty"`      // Naturalidade da pessoa no padrão do DIMOB.
	CodNaturalidade   *int                   `json:"CodNaturalidade,omitempty"`   // Naturalidade da pessoa no DIMOB.
	Contato           *string                `json:"Contato,omitempty"`           // Informações de pessoa de contato.
	CodIntegracaoSist *string                `json:"CodIntegracaoSist,omitempty"` // Código de integração/migração de sistema.
	Sexo              *string                `json:"Sexo,omitempty"`              // Sexo/gênero da pessoa. Valor default é ' '.
	TipoPessoa        *string                `json:"TipoPessoa,omitempty"`        // Tipo da pessoa. Valor default é ' '.
	CpfCnpj           *string                `json:"CpfCnpj,omitempty"`           // Se for tipo de pessoa física o valor é um CPF. Se for tipo de pessoa jurídica o valor é um CNPJ. Se o tipo de pessoa não for informado então este campo é vazio.
	RG                *string                `json:"RG,omitempty"`                // Número do documento de identificação da pessoa física. Não preencher se for pessoa jurídica.
	OrgaoExpedidor    *string                `json:"OrgaoExpedidor,omitempty"`    // Órgão que expediu o documento de identificação informado.
	DataExpedicao     *string                `json:"DataExpedicao,omitempty"`     // A data de expedição do documento de identificação informado.
	DataNascimento    *string                `json:"DataNascimento,omitempty"`    // Data de nascimento da pessoa física ou de criação da pessoa jurídica.
	CodConjuge        *int                   `json:"CodConjuge,omitempty"`        // Código de pessoa do cônjuge.
	SenhaInternet     *string                `json:"SenhaInternet,omitempty"`     // Senha de acesso no site/internet.
	Email             *string                `json:"Email,omitempty"`             // E-mail da pessoa.
	TipoEnderCobr     *string                `json:"TipoEnderCobr,omitempty"`     // Tipo de endereço de cobrança que deve existir no array 'Enderecos'.
	TipoEnderCorresp  *string                `json:"TipoEnderCorresp,omitempty"`  // Tipo de endereço de correpondência que deve existir no array 'Enderecos'.
	Passaporte        *string                `json:"Passaporte,omitempty"`        // Número do passaporte da pessoa física.
	Celular           *string                `json:"Celular,omitempty"`           // Número de celular.
	TipoConta         *string                `json:"TipoConta,omitempty"`         // *Tipo da conta bancária desta pessoa.
	CodBanco          *int                   `json:"CodBanco,omitempty"`          // *Código do banco.
	CodAgencia        *int                   `json:"CodAgencia,omitempty"`        // *Código da agência bancária.
	ContaCorrente     *string                `json:"ContaCorrente,omitempty"`     // *Número da conta corrente desta pessoa.
	Classificacao     *string                `json:"Classificacao,omitempty"`     // Código de classificacão desta pessoa.
	Observacao        *string                `json:"Observacao,omitempty"`        // Texto de observação desta pessoa.
	EstadoCivil       *string                `json:"EstadoCivil,omitempty"`       // Estado civil da pessoa.
	CodProfissao      *int                   `json:"CodProfissao,omitempty"`      // Código da profissão desta pessoa.
	Ativo             *string                `json:"Ativo,omitempty"`             // Indica se está ativo.
	EmailAutomatico   *string                `json:"EmailAutomatico,omitempty"`   // Avisos automáticos por e-mail.
	EmailNfse         *string                `json:"EmailNfse,omitempty"`         // Utilizado na emissão na NFSe.
	WhatsPrioritario  *string                `json:"WhatsPrioritario,omitempty"`  // Campanhas ativas por WhatsApp.
	Enderecos         *[]ActionInputEndereco `json:"Enderecos,omitempty"`         // A pessoa pode ter mais de um endereço, sendo um residencial outro comercial, etc.
}

type ActionInputEndereco struct {
	TipoEnder   string `json:"TipoEnder,omitempty"`   // *Tipo de endereço.
	CEP         int    `json:"CEP,omitempty"`         // *Número do CEP.
	TipoLograd  string `json:"TipoLograd,omitempty"`  // Tipo de logradouro abreviado ou por extenso ('R' ou 'RUA', 'AV' ou 'AVENIDA', etc.).
	Logradouro  string `json:"Logradouro,omitempty"`  // *Logradouro do endereço. Deve ser informado apenas o nome sem o tipo de logradouro.
	Numero      int    `json:"Numero,omitempty"`      // Número do endereço.
	Complemento string `json:"Complemento,omitempty"` // Complemento do endereço.
	Bairro      string `json:"Bairro,omitempty"`      // *Bairro do endereço.
	Cidade      string `json:"Cidade,omitempty"`      // *Cidade do endereço.
	UF          string `json:"UF,omitempty"`          // *Sigla da Unidade Federativa do endereço.
	Telefone1   string `json:"Telefone1,omitempty"`   // Número de telefone principal.
	Telefone2   string `json:"Telefone2,omitempty"`   // Número de telefone alternativo.
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
	CodPessoa *int `json:"CodPessoa,omitempty"` // Código da pessoa.
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
