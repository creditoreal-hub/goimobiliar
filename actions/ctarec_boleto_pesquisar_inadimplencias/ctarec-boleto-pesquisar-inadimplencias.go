package ctarec_boleto_pesquisar_inadimplencias

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

var ACTION = "CTAREC_BOLETO_PESQUISAR_INADIMPLENCIAS"

type ActionInput struct {
	IniRel                   *string  `json:"IniRel,omitempty"`                   //
	OrigemCondom             *string  `json:"OrigemCondom,omitempty"`             //
	ExcetoQuitSeg            *string  `json:"ExcetoQuitSeg,omitempty"`            //
	CodIndiceCorr            *string  `json:"CodIndiceCorr,omitempty"`            //
	TipoRel                  *int     `json:"TipoRel,omitempty"`                  // Valor default é '0'.
	Retidos                  *string  `json:"Retidos,omitempty"`                  //
	DataInicial              *string  `json:"DataInicial,omitempty"`              // *Formato DD/MM/YYYY.
	DataFinal                *string  `json:"DataFinal,omitempty"`                // *Formato DD/MM/YYYY.
	CodInicial               *string  `json:"CodInicial,omitempty"`               //
	CodFinal                 *int     `json:"CodFinal,omitempty"`                 //
	IdEconomia               *int     `json:"IdEconomia,omitempty"`               // Chave principal da economia/unidade.
	CodFilial                *string  `json:"CodFilial,omitempty"`                //
	TipoFianca               *string  `json:"TipoFianca,omitempty"`               //
	DataBase                 *string  `json:"DataBase,omitempty"`                 // Formato DD/MM/YYYY.
	RetInadInicial           *string  `json:"RetInadInicial,omitempty"`           // Formato DD/MM/YYYY.
	Classificacao            *string  `json:"Classificacao,omitempty"`            //
	ExportaSindico           *string  `json:"ExportaSindico,omitempty"`           //
	RelPorImov               *string  `json:"RelPorImov,omitempty"`               //
	SemQuitaAposVencFinal    *string  `json:"SemQuitaAposVencFinal,omitempty"`    //
	ApenasRetInad            *string  `json:"ApenasRetInad,omitempty"`            //
	LancamentoAnalitico      *string  `json:"LancamentoAnalitico,omitempty"`      //
	SemTaxasSemMulta         *string  `json:"SemTaxasSemMulta,omitempty"`         //
	DebConta                 *string  `json:"DebConta,omitempty"`                 //
	DocsAcordo               *string  `json:"DocsAcordo,omitempty"`               //
	ExcetoDocsAcordo         *string  `json:"ExcetoDocsAcordo,omitempty"`         //
	IncluirDocsAcordo        *string  `json:"IncluirDocsAcordo,omitempty"`        //
	OrdemEnd                 *string  `json:"OrdemEnd,omitempty"`                 //
	InformaFone              *string  `json:"InformaFone,omitempty"`              //
	CondObsJur               *string  `json:"CondObsJur,omitempty"`               //
	ObsJurAcoes              *string  `json:"ObsJurAcoes,omitempty"`              //
	ObsJurProc               *string  `json:"ObsJurProc,omitempty"`               //
	ExcetoGarantidos         *string  `json:"ExcetoGarantidos,omitempty"`         //
	ApenasGarantidos         *string  `json:"ApenasGarantidos,omitempty"`         //
	ApenasProgramados        *string  `json:"ApenasProgramados,omitempty"`        //
	ApenasCondominioAtivo    *string  `json:"ApenasCondominioAtivo,omitempty"`    //
	ApenasCondominioInativo  *string  `json:"ApenasCondominioInativo,omitempty"`  //
	PercentualHonorarios     *float64 `json:"PercentualHonorarios,omitempty"`     // *
	TemCustas                *string  `json:"TemCustas,omitempty"`                //
	CodBloco                 *string  `json:"CodBloco,omitempty"`                 // Código do bloco da economia.
	CodAdvogado              *int     `json:"CodAdvogado,omitempty"`              // Código do Advogado.
	Ocupados                 *string  `json:"Ocupados,omitempty"`                 //
	Boletos                  *string  `json:"Boletos,omitempty"`                  //
	VlrHonorarios            *float64 `json:"VlrHonorarios,omitempty"`            //
	CodAssessor              *int     `json:"CodAssessor,omitempty"`              //
	ExibirParcelamentoAcordo *string  `json:"ExibirParcelamentoAcordo,omitempty"` //
	ApenasComAdv             *string  `json:"ApenasComAdv,omitempty"`             //
	AnaliticoEstorno         *string  `json:"AnaliticoEstorno,omitempty"`         //
	ApenasEconAtivas         *string  `json:"ApenasEconAtivas,omitempty"`         //
	Competencia              *string  `json:"Competencia,omitempty"`              // Competência do documento no formato 'YYYYMM'.
	ExibirPercentualInad     *string  `json:"ExibirPercentualInad,omitempty"`     //
	Desocupados              *string  `json:"Desocupados,omitempty"`              //
	CodLocatario             *int     `json:"CodLocatario,omitempty"`             //
	CodFornecedorAdm         *int     `json:"CodFornecedorAdm,omitempty"`         //
	TotTxAdm                 *string  `json:"TotTxAdm,omitempty"`                 //
	Inativos                 *string  `json:"Inativos,omitempty"`                 //
	ExibirAgrupados          *string  `json:"ExibirAgrupados,omitempty"`          //
	ApenasSemAdv             *string  `json:"ApenasSemAdv,omitempty"`             //
	QtdeLinhas               *int     `json:"QtdeLinhas,omitempty"`               // Quantidade máxima de linhas de resposta, utilizado para obter resultados por segmentos (paginação). Se não for informado então a resposta conterá todas as linhas selecionadas pela ação. Valor default é '0'.
	ProximasLinhas           *string  `json:"ProximasLinhas,omitempty"`           // Campo opcional indicando que, ao invés de executar a ação, solicita as linhas do próximo segmento. Valor default é 'N'.
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
	Pendentes *[]RequestResponseBodyPendente `json:"Pendentes,omitempty"` //
	Acordo    *string                        `json:"acordo,omitempty"`    //
}

type RequestResponseBodyPendente struct {
	Bairro                  *string  `json:"Bairro,omitempty"`                  // Bairro do endereço.
	Cidade                  *string  `json:"Cidade,omitempty"`                  // Cidade do endereço.
	UF                      *string  `json:"UF,omitempty"`                      // Sigla da Unidade Federativa do endereço.
	CEP                     *string  `json:"CEP,omitempty"`                     // Número do CEP.
	TipoPessoa              *string  `json:"TipoPessoa,omitempty"`              // Tipo da pessoa.
	CpfCnpj                 *int     `json:"CpfCnpj,omitempty"`                 // Se for tipo de pessoa física o valor é um CPF. Se for tipo de pessoa jurídica o valor é um CNPJ. Se o tipo de pessoa não for informado então este campo é vazio.
	Email                   *string  `json:"Email,omitempty"`                   // E-mail da pessoa.
	Telefone                *string  `json:"Telefone,omitempty"`                // Número do CEP.
	Nome                    *string  `json:"Nome,omitempty"`                    // Nome da pessoa.
	CodPessoa               *int     `json:"CodPessoa,omitempty"`               // Código de pessoa do sacado.
	DataVencimento          *string  `json:"DataVencimento,omitempty"`          // Data de vencimento do lançamento.
	Competencia             *string  `json:"Competencia,omitempty"`             // Competência do documento no formato 'YYYYMM'.
	DataCartaInadimplencia1 *string  `json:"DataCartaInadimplencia1,omitempty"` // Data carta inadimplencia 1.
	DataCartaInadimplencia2 *string  `json:"DataCartaInadimplencia2,omitempty"` // Data carta inadimplencia 2.
	DataCartaInadimplencia3 *string  `json:"DataCartaInadimplencia3,omitempty"` // Data carta inadimplencia 3.
	DataJuridico            *string  `json:"DataJuridico,omitempty"`            // Data ida para juridico.
	TipoDocumento           *string  `json:"TipoDocumento,omitempty"`           //
	Nossonumero             *string  `json:"Nossonumero,omitempty"`             // Número de identificação bancário.
	DocCapaId               *int     `json:"DocCapaId,omitempty"`               // Código interno do boleto (seu código).
	DataGeracao             *string  `json:"DataGeracao,omitempty"`             // Data geração.
	BaseJuro                *string  `json:"BaseJuro,omitempty"`                // Tipo de cobrança de juros.
	CodFilial               *string  `json:"CodFilial,omitempty"`               //
	PercJuros               *float64 `json:"PercJuros,omitempty"`               // Percentual de juros em caso de atraso de pagamento.
	PercMulta               *float64 `json:"PercMulta,omitempty"`               // Percentual de multa.
	VlrTaxaPorte            *float64 `json:"VlrTaxaPorte,omitempty"`            // 	Valor da taxa porte.
	MsgCalcCorrecao         *string  `json:"MsgCalcCorrecao,omitempty"`         //
	CodCondominio           *int     `json:"CodCondominio,omitempty"`           // Código do condomínio.
	UsuarioId               *string  `json:"UsuarioId,omitempty"`               // Usuário que registrou observação.
	Economia                *string  `json:"Economia,omitempty"`                //
	IdEconomia              *int     `json:"IdEconomia,omitempty"`              // Chave principal da economia/unidade.
	CodBloco                *string  `json:"CodBloco,omitempty"`                // Código do bloco da economia.
	CodBlocoLancto          *string  `json:"CodBlocoLancto,omitempty"`          // Código do bloco do lançamento.
	CodImovel               *int     `json:"CodImovel,omitempty"`               // Código do imóvel.
	ExportaLocacao          *string  `json:"ExportaLocacao,omitempty"`          // Indica se exporta para locação.
	DescrClasseImovel       *string  `json:"DescrClasseImovel,omitempty"`       // Descrição da classe de imóvel da economia/unidade.
	NomeCondominio          *string  `json:"NomeCondominio,omitempty"`          // Nome do condomínio.
	ValorJuros              *float64 `json:"ValorJuros,omitempty"`              // Valor dos juros.
	Correcao                *float64 `json:"Correcao,omitempty"`                // Correção monetária sobre valor original.
	VlrDocumento            *float64 `json:"VlrDocumento,omitempty"`            // Valor do documento.
	Sexo                    *string  `json:"Sexo,omitempty"`                    // Sexo/gênero da pessoa.
	DataVencFianca          *string  `json:"DataVencFianca,omitempty"`          // Data de vencimento do seguro fiança.
	DataVigInicial          *string  `json:"DataVigInicial,omitempty"`          // Data inicial da vigência do contrato.
	DataDistrato            *string  `json:"DataDistrato,omitempty"`            // Data de encerramento.
	CodContratoLoc          *int     `json:"CodContratoLoc,omitempty"`          // Código do contrato de locação deste imóvel.
	Endereco                *string  `json:"Endereco,omitempty"`                // Endereço do condomínio.
	TipoFianca              *string  `json:"TipoFianca,omitempty"`              //
	DiaPagtoProp            *int     `json:"DiaPagtoProp,omitempty"`            // Dia do mês para o pagamento ao proprietário quando a forma de cálculo for 'Programado'.
	CodTaxa                 *int     `json:"CodTaxa,omitempty"`                 // Código da taxa que classifica este lançamento.
	DescricaoTaxa           *string  `json:"DescricaoTaxa,omitempty"`           // Descricao da Taxa.
	VlrLancamento           *float64 `json:"VlrLancamento,omitempty"`           // Valor original do Lançamento Analítico.
	JurosLancamento         *float64 `json:"JurosLancamento,omitempty"`         // Valor dos juros do Lançamento Analítico.
	MultaLancamento         *float64 `json:"MultaLancamento,omitempty"`         // Valor da Multa do Lançamento Analítico.
	CorrecaoLancamento      *float64 `json:"CorrecaoLancamento,omitempty"`      // Correção monetária sobre valor original do Lançamento Analítico.
	AdvogadoBoleto          *string  `json:"AdvogadoBoleto,omitempty"`          // Nome do Advogado no Boleto.
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
