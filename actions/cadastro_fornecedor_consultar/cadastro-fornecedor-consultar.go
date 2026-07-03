package cadastro_fornecedor_consultar

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

var ACTION = "CADASTRO_FORNECEDOR_CONSULTAR"

type ActionInput struct {
	CodFornecedor *string `json:"CodFornecedor,omitempty"` // Código do fornecedor.
	CpfCnpj       *string `json:"CpfCnpj,omitempty"`       // CPF ou CNPJ do fornecedor.
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
	CodFornecedor       *int           `json:"CodFornecedor,omitempty"`       // Código do fornecedor.
	Nome                *string        `json:"Nome,omitempty"`                // Nome/Razão Social do fornecedor.
	NomeFantasia        *string        `json:"NomeFantasia,omitempty"`        // Nome de fantasia do fornecedor.
	TipoPessoa          *string        `json:"TipoPessoa,omitempty"`          // Tipo de pessoa do fornecedor.
	CpfCnpj             *types.CpfCnpj `json:"CpfCnpj,omitempty"`             // CPF ou CNPJ do fornecedor.
	InscricaoInss       *string        `json:"InscricaoInss,omitempty"`       // CPF/CNPJ do fornecedor.
	InscricaoMunicipal  *string        `json:"InscricaoMunicipal,omitempty"`  // Inscrição municipal do fornecedor.
	Categoria           *string        `json:"Categoria,omitempty"`           // Categoria do fornecedor.
	PIS                 *string        `json:"PIS,omitempty"`                 // PIS do fornecedor.
	TipoConta           *string        `json:"TipoConta,omitempty"`           // Tipo da conta bancária do fornecedor.
	CodBanco            *int           `json:"CodBanco,omitempty"`            // Código do banco.
	CodAgencia          *int           `json:"CodAgencia,omitempty"`          // Código da agência bancária.
	ContaCorrente       *string        `json:"ContaCorrente,omitempty"`       // Número da conta corrente do fornecedor.
	Contato             *string        `json:"Contato,omitempty"`             // Contato no fornecedor.
	CargoContato        *string        `json:"CargoContato,omitempty"`        // Cargo do contato no fornecedor.
	CEP                 *int           `json:"CEP,omitempty"`                 // Número do CEP.
	TipoLograd          *string        `json:"TipoLograd,omitempty"`          // Tipo de logradouro abreviado ou por extenso ('R' ou 'RUA', 'AV' ou 'AVENIDA', etc.).
	Logradouro          *string        `json:"Logradouro,omitempty"`          // Logradouro do endereço. Deve ser informado apenas o nome sem o tipo de logradouro.
	Numero              *int           `json:"Numero,omitempty"`              // Número do endereço.
	Complemento         *string        `json:"Complemento,omitempty"`         // Complemento do endereço.
	Bairro              *string        `json:"Bairro,omitempty"`              // Bairro do endereço.
	Cidade              *string        `json:"Cidade,omitempty"`              // Cidade do endereço.
	UF                  *string        `json:"UF,omitempty"`                  // Sigla da Unidade Federativa do endereço.
	Telefone1           *string        `json:"Telefone1,omitempty"`           // Número do telefone principal.
	Celular             *string        `json:"Celular,omitempty"`             // Número do celular do fornecedor.
	Email               *string        `json:"Email,omitempty"`               // E-mail do fornecedor.
	FormaPagamento      *string        `json:"FormaPagamento,omitempty"`      // Forma de pagamento do fornecedor.
	TipoChavePix        *string        `json:"TipoChavePix,omitempty"`        // Tipo da chave PIX.
	ChavePix            *string        `json:"ChavePix,omitempty"`            // Chave PIX.
	TipoDocumento       *string        `json:"TipoDocumento,omitempty"`       // Tipos de documentos.
	EmiteNFSE           *string        `json:"EmiteNFSE,omitempty"`           // Indica se fornecedor emite NFSe.
	Ativo               *string        `json:"Ativo,omitempty"`               // Indica se está ativo.
	CodPessoaFavorecido *int           `json:"CodPessoaFavorecido,omitempty"` // Código da pessoa favorecida em pagamentos ao fornecedor.
	Favorecido          *string        `json:"Favorecido,omitempty"`          // Nome da pessoa favorecida.
	CodPessoaTitular    *int           `json:"CodPessoaTitular,omitempty"`    // Código da pessoa titular da empresa para fins previdenciários.
	Titular             *string        `json:"Titular,omitempty"`             // Nome da pessoa titular da empresa para fins previdenciários.
	MEI                 *string        `json:"MEI,omitempty"`                 // MEI do fornecedor.
	NIT                 *string        `json:"NIT,omitempty"`                 // NIT do fornecedor.
	ProdutorRural       *string        `json:"ProdutorRural,omitempty"`       // Indica se o fornecedor é produtor rural.
	CodigoCBO           *string        `json:"CodigoCBO,omitempty"`           // Código CBO (Classificação Brasileira de Ocupações).
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
