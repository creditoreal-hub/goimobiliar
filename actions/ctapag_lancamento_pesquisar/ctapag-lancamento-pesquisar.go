package ctapag_lancamento_pesquisar

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

var ACTION = "CTAPAG_LANCAMENTO_PESQUISAR"

type ActionInput struct {
	TipoPesquisa          *string  `json:"TipoPesquisa,omitempty"`          // *Indica o tipo de pesquisa/origem.
	CodCondominio         *int     `json:"CodCondominio,omitempty"`         // Código do condomínio do lançamento (se origem for 'C').
	CodImovel             *int     `json:"CodImovel,omitempty"`             // Código do imóvel do lançamento (se origem for 'I').
	CodPessoaProprietario *int     `json:"CodPessoaProprietario,omitempty"` // Código do proprietário.
	CodPlanoContaAdm      *int     `json:"CodPlanoContaAdm,omitempty"`      // Código da conta no plano de contas da administradora (se origem for 'A').
	CodBloco              *string  `json:"CodBloco,omitempty"`              // Código do bloco do lançamento (se origem for 'C').
	CodFornecedor         *int     `json:"CodFornecedor,omitempty"`         // Código do fornecedor do lançamento.
	CodPessoaFavorecido   *int     `json:"CodPessoaFavorecido,omitempty"`   // Código do favorecido no cadastro de pessoas.
	NomeFavorecido        *string  `json:"NomeFavorecido,omitempty"`        // Nome do favorecido.
	GrupoSoma             *int     `json:"GrupoSoma,omitempty"`             // Código do grupo de soma.
	CodTaxa               *int     `json:"CodTaxa,omitempty"`               // Código da taxa que classifica este lançamento.
	TipoPeriodo           *string  `json:"TipoPeriodo,omitempty"`           // Indica o tipo de período a ser pesquisado. Valor default é 'V'.
	DataInicial           *string  `json:"DataInicial,omitempty"`           // Primeiro dia do período a ser pesquisado.
	DataFinal             *string  `json:"DataFinal,omitempty"`             // Último dia do período a ser pesquisado.
	Competencia           *string  `json:"Competencia,omitempty"`           // Competência do lançamento no formato 'YYYYMM'.
	Status                *string  `json:"Status,omitempty"`                // Indica a situação dos lançamentos a serem pesquisados. Valor default é 'T'.
	PrevisaoReal          *string  `json:"PrevisaoReal,omitempty"`          // Indica o tipo dos lançamentos a serem pesquisados. Valor default é 'T'.
	NumeroDocumento       *string  `json:"NumeroDocumento,omitempty"`       // Número do documento do fornecedor.
	UsuarioInclusao       *string  `json:"UsuarioInclusao,omitempty"`       // Usuário que incluiu o lançamento.
	ValorLiquido          *types.Float64 `json:"ValorLiquido,omitempty"`          // Valor líquido do lançamento. Valor default é '0'.
	ValorBruto            *types.Float64 `json:"ValorBruto,omitempty"`            // Valor bruto do lançamento. Valor default é '0'.
	ValorPagamento        *types.Float64 `json:"ValorPagamento,omitempty"`        // Valor do pagamento. Valor default é '0'.
	QtdeLinhas            *int     `json:"QtdeLinhas,omitempty"`            // Quantidade máxima de linhas de resposta, utilizado para obter resultados por segmentos (paginação). Se não for informado então a resposta conterá todas as linhas selecionadas pela ação. Valor default é '0'.
	ProximasLinhas        *string  `json:"ProximasLinhas,omitempty"`        // Campo opcional indicando que, ao invés de executar a ação, solicita as linhas do próximo segmento. Valor default é 'N'.
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
	Lancamentos *[]RequestResponseBodyLancamento `json:"Lancamentos,omitempty"`
}

type RequestResponseBodyLancamento struct {
	NumeroLancto     *int     `json:"NumeroLancto,omitempty"`     // Número do lançamento.
	DataVencimento   *string  `json:"DataVencimento,omitempty"`   // Data de vencimento do lançamento.
	CodTaxa          *int     `json:"CodTaxa,omitempty"`          // Código da taxa que classifica este lançamento.
	DescrTaxa        *string  `json:"DescrTaxa,omitempty"`        // Descrição da taxa que classifica este lançamento.
	NomeFavorecido   *string  `json:"NomeFavorecido,omitempty"`   // Nome do favorecido.
	ValorLiquido     *types.Float64 `json:"ValorLiquido,omitempty"`     // Valor líquido do lançamento.
	PrevisaoReal     *string  `json:"PrevisaoReal,omitempty"`     // Indica o tipo dos lançamentos a serem pesquisados.
	Frequencia       *string  `json:"Frequencia,omitempty"`       // Define se lançamento é único ou permanente.
	Pago             *string  `json:"Pago,omitempty"`             // Indica se o lançamento está pago.
	NumeroDocumento  *string  `json:"NumeroDocumento,omitempty"`  // Número do documento do fornecedor.
	UsuarioSuspensao *string  `json:"UsuarioSuspensao,omitempty"` // Usuário que suspendeu o lançamento.
	DataSuspensao    *string  `json:"DataSuspensao,omitempty"`    // Data da suspensão do lançamento.
	MotivoSuspensao  *string  `json:"MotivoSuspensao,omitempty"`  // Motivo da suspensão do lançamento.
	ValorPagamento   *types.Float64 `json:"ValorPagamento,omitempty"`   // Valor do pagamento.
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
