package locacao_imovel_alterar

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

var ACTION = "LOCACAO_IMOVEL_ALTERAR"

type ActionInput struct {
	CodImovel                    *int                         `json:"CodImovel,omitempty"`                    // *Código do imóvel.
	TipoLograd                   *string                      `json:"TipoLograd,omitempty"`                   // Abreviatura do tipo de logradouro ('R', 'AV', etc.).
	Logradouro                   *string                      `json:"Logradouro,omitempty"`                   // Logradouro do endereço.
	Numero                       *int                         `json:"Numero,omitempty"`                       // Número do endereço.
	Complemento                  *string                      `json:"Complemento,omitempty"`                  // Complemento do endereço.
	CEP                          *string                      `json:"CEP,omitempty"`                          // Número do CEP.
	Bairro                       *string                      `json:"Bairro,omitempty"`                       // Bairro do endereço.
	Cidade                       *string                      `json:"Cidade,omitempty"`                       // Cidade do endereço.
	UF                           *string                      `json:"UF,omitempty"`                           // Sigla da Unidade Federativa do endereço.
	CodClasseImovel              *int                         `json:"CodClasseImovel,omitempty"`              // Código da classe do imóvel.
	CodAssessor                  *int                         `json:"CodAssessor,omitempty"`                  // Código do assessor/gestor.
	Telefone                     *string                      `json:"Telefone,omitempty"`                     // Número do CEP.
	DataInclusao                 *string                      `json:"DataInclusao,omitempty"`                 // Data de inclusão do imóvel.
	Situacao                     *string                      `json:"Situacao,omitempty"`                     // Indica a situação do imóvel.
	Ativo                        *string                      `json:"Ativo,omitempty"`                        // Indica se está ativo.
	QtdeDormitorios              *string                      `json:"QtdeDormitorios,omitempty"`              // Quantidade de dormitórios.
	QtdeGaragem                  *string                      `json:"QtdeGaragem,omitempty"`                  // Quantidade de vagas de garagem.
	AreaTotal                    *string                      `json:"AreaTotal,omitempty"`                    // Área total do imóvel.
	AreaPrivativa                *string                      `json:"AreaPrivativa,omitempty"`                // Área privativa do imóvel.
	Venda                        *string                      `json:"Venda,omitempty"`                        // Indica se o imóvel é para oferta de venda.
	Locacao                      *string                      `json:"Locacao,omitempty"`                      // Indica se o imóvel é para oferta de locação.
	MesesGarantiaAlug            *string                      `json:"MesesGarantiaAlug,omitempty"`            // Número de meses de garantia do aluguel.
	MesesGarantiaEnc             *string                      `json:"MesesGarantiaEnc,omitempty"`             // Número de meses de garantia dos encargos.
	Matricula                    *string                      `json:"Matricula,omitempty"`                    // Matrícula do imóvel.
	ZonaRegistro                 *string                      `json:"ZonaRegistro,omitempty"`                 // Zona do Cartório de Registro do imóvel.
	ValorVenda                   *string                      `json:"ValorVenda,omitempty"`                   // Valor de venda do imóvel.
	NomePredio                   *string                      `json:"NomePredio,omitempty"`                   // Nome do prédio do imóvel.
	Latitude                     *string                      `json:"Latitude,omitempty"`                     // Latitude do imóvel em graus e decimais do grau.
	Longitude                    *string                      `json:"Longitude,omitempty"`                    // Longitude do imóvel em graus e decimais do grau.
	ValorAluguel                 *string                      `json:"ValorAluguel,omitempty"`                 // Valor de aluguel do imóvel.
	Imediacao                    *string                      `json:"Imediacao,omitempty"`                    // Descrição das imediações do imóvel.
	DescrCaracteristicas         *string                      `json:"DescrCaracteristicas,omitempty"`         // Descrição das características do imóvel.
	DescrReduzida                *string                      `json:"DescrReduzida,omitempty"`                // Descrição reduzida do imóvel.
	ObsIPTU                      *string                      `json:"ObsIPTU,omitempty"`                      // Observações referentes ao IPTU.
	ObsJuridico                  *string                      `json:"ObsJuridico,omitempty"`                  // Observações referentes a atividades jurídicas.
	ObsAcoes                     *string                      `json:"ObsAcoes,omitempty"`                     // Observações referentes a ações feitas ou a fazer no imóvel.
	ObsSeguros                   *string                      `json:"ObsSeguros,omitempty"`                   // Observações referentes ao seguro.
	ObsCadastro                  *string                      `json:"ObsCadastro,omitempty"`                  // Observações referentes ao cadastro do imóvel.
	ObsCtaPagar                  *string                      `json:"ObsCtaPagar,omitempty"`                  // Observações referentes a contas a pagar.
	ObsDOC                       *string                      `json:"ObsDOC,omitempty"`                       // Observações referentes ao cálculo do DOC.
	ObsTaxasCond                 *string                      `json:"ObsTaxasCond,omitempty"`                 // Observações referentes às taxas de condomínio.
	LoginAdmCondom               *string                      `json:"LoginAdmCondom,omitempty"`               // Login de acesso as administradoras de condomínio.
	SenhaAdmCondom               *string                      `json:"SenhaAdmCondom,omitempty"`               // Senha de acesso as administradoras de condomínio. OBSERVAÇÃO: Para fins de segurança, a senha informada neste campo vem criptografada e deve ser um tratamento específico. Ao invés de ser comparada diretamente com a senha digitada pelo usuário, a senha digitada deve ser convertida para maiúsculo e então criptografada em MD5. O valor obtido em MD5 é que deve ser usada na comparação. Exemplo em pseudo-linguagem:
	ObsOutras                    *string                      `json:"ObsOutras,omitempty"`                    // Observações gerais.
	ObsInternet                  *string                      `json:"ObsInternet,omitempty"`                  // Observações que devem ser enviadas para o site na internet.
	ValorCondominio              *float64                     `json:"ValorCondominio,omitempty"`              // Valor mensal do condomínio do imóvel.
	ValorIPTU                    *float64                     `json:"ValorIPTU,omitempty"`                    // Valor mensal de IPTU do imóvel.
	NroInscricaoIPTU             *int                         `json:"NroInscricaoIPTU,omitempty"`             // Número de inscrição do IPTU.
	InformativoDOC               *string                      `json:"InformativoDOC,omitempty"`               // Texto que deve constar na área do informativo do DOC.
	InstrucaoDOC                 *string                      `json:"InstrucaoDOC,omitempty"`                 // Texto que deve constar na área de instruções do DOC.
	IncideIRFTxAdm               *string                      `json:"IncideIRFTxAdm,omitempty"`               // Indica se incide imposto de renda sobre a taxa de administração.
	FormaCalcPagto               *string                      `json:"FormaCalcPagto,omitempty"`               // Indica a forma de cálculo para o pagamento ao proprietário.
	TaxaIntermediacao            *float64                     `json:"TaxaIntermediacao,omitempty"`            // Percentual da taxa de intermediação.
	IncidenciaTaxaAdm            *string                      `json:"IncidenciaTaxaAdm,omitempty"`            // Incidência da taxa de administração.
	TaxaAdm                      *float64                     `json:"TaxaAdm,omitempty"`                      // Taxa de administração do imóvel em forma de um percentual sobre o aluguel. Se for um valor fixo em Reais então informá-lo no campo 'ValorTaxaAdm' mas apenas um deles deve ser informado.
	ValorTaxaAdm                 *float64                     `json:"ValorTaxaAdm,omitempty"`                 // Taxa de administração do imóvel em forma de um valor fixo em Reais. Se for um percentual sobre o aluguel então informá-lo no campo 'TaxaAdm' mas apenas um deles deve ser informado.
	IncidenciaValorMinimoTaxaAdm *string                      `json:"IncidenciaValorMinimoTaxaAdm,omitempty"` // Indicação de cláusula de valor mínimo de taxa de administração.
	ValorMinimoTaxaAdm           *float64                     `json:"ValorMinimoTaxaAdm,omitempty"`           // Valor mínimo de taxa de administração quando indicado 'Cláusula de valor mínimo de taxa de administração' (IncideValorMinimoTaxaAdm).
	CobrancaAntecipada           *string                      `json:"CobrancaAntecipada,omitempty"`           // Indica se tem desconto de pontualidade quando pago antes do vencimento.
	RamalAgua                    *string                      `json:"RamalAgua,omitempty"`                    // Identificação do ramal/registro de água.
	CodAgencia                   *int                         `json:"CodAgencia,omitempty"`                   // Código da agência/loja de captação do imóvel.
	OrigemCaptacao               *string                      `json:"OrigemCaptacao,omitempty"`               // Descrição da origem da captação do imóvel.
	CodAgenciador                *int                         `json:"CodAgenciador,omitempty"`                // Código do agenciador de captação do imóvel.
	CodFornecCond                *int                         `json:"CodFornecCond,omitempty"`                // Código de fornecedor da administradora de condomínio.
	GrupoAnalise                 *int                         `json:"GrupoAnalise,omitempty"`                 // Grupo de análise.
	GaranteAluguel               *string                      `json:"GaranteAluguel,omitempty"`               // Indica se tem garantia de aluguel.
	GaranteEncargos              *string                      `json:"GaranteEncargos,omitempty"`              // Indica se tem garantia de encargos.
	CobraDiferencaCond           *string                      `json:"CobraDiferencaCond,omitempty"`           // Indica se cobra diferença de condomínio.
	NomeLocat                    *string                      `json:"NomeLocat,omitempty"`                    // Nome do locatário.
	CodIntegracaoSist            *string                      `json:"CodIntegracaoSist,omitempty"`            // Código deste imóvel no sistema integrado/migrado.
	VencMesComp                  *string                      `json:"VencMesComp,omitempty"`                  // Indica vencimento no mês de competência.
	Caracteristicas              *[]ActionInputCaracteristica `json:"Caracteristicas,omitempty"`              // ATENÇÃO: todas as características existentes serão substituidas pelas que forem informadas neste array. Ou seja, não enviar este array para manter as características atuais. Se este array for enviado então deve constar o conjunto completo de características do imóvel.
}

type ActionInputCaracteristica struct {
	CodCaract *int `json:"CodCaract,omitempty"` // *Código da característica do imóvel.
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
	CodImovel *int `json:"CodImovel,omitempty"` // Código do imóvel.
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
