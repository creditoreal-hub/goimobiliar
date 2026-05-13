package locacao_imovel_incluir

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

var ACTION = "LOCACAO_IMOVEL_INCLUIR"

type ActionInput struct {
	TipoLograd                   *string                      `json:"TipoLograd,omitempty"`                   //	String(10)	Abreviatura do tipo de logradouro ('R', 'AV', etc.).
	Logradouro                   *string                      `json:"Logradouro,omitempty"`                   //	String(60)	*Logradouro do endereço.
	Numero                       *int                         `json:"Numero,omitempty"`                       //	Number(5)	*Número do endereço.
	Complemento                  *string                      `json:"Complemento,omitempty"`                  //	String(20)	*Complemento do endereço.
	CEP                          *string                      `json:"CEP,omitempty"`                          //	String(8)	*Número do CEP.
	Bairro                       *string                      `json:"Bairro,omitempty"`                       //	String(60)	*Bairro do endereço.
	Cidade                       *string                      `json:"Cidade,omitempty"`                       //	String(40)	*Cidade do endereço.
	UF                           *string                      `json:"UF,omitempty"`                           //	String(2)	*Sigla da Unidade Federativa do endereço.
	CodClasseImovel              *int                         `json:"CodClasseImovel,omitempty"`              //	Number(2)	*Código da classe do imóvel.
	CodAssessor                  *int                         `json:"CodAssessor,omitempty"`                  //	Number(4)	Código do assessor/gestor.
	Telefone                     *string                      `json:"Telefone,omitempty"`                     //	Phone(20)	Número do CEP.
	DataInclusao                 *string                      `json:"DataInclusao,omitempty"`                 //	Date	Data de inclusão do imóvel.
	Situacao                     *string                      `json:"Situacao,omitempty"`                     //	String(3)	Indica a situação do imóvel. Valor default é 'NOR'.
	Ativo                        *string                      `json:"Ativo,omitempty"`                        //	String(1)	Indica se está ativo. Valor default é 'S'.
	QtdeDormitorios              *int                         `json:"QtdeDormitorios,omitempty"`              //	Number(2)	Quantidade de dormitórios. Valor default é '0'.
	QtdeGaragem                  *int                         `json:"QtdeGaragem,omitempty"`                  //	Number(3)	Quantidade de vagas de garagem. Valor default é '0'.
	AreaTotal                    *types.Float64                     `json:"AreaTotal,omitempty"`                    //	Number(14,2)	Área total do imóvel.
	AreaPrivativa                *types.Float64                     `json:"AreaPrivativa,omitempty"`                //	Number(14,2)	Área privativa do imóvel.
	Venda                        *string                      `json:"Venda,omitempty"`                        //	String(1)	Indica se o imóvel é para oferta de venda.
	Locacao                      *string                      `json:"Locacao,omitempty"`                      //	String(1)	Indica se o imóvel é para oferta de locação.
	MesesGarantiaAlug            *int                         `json:"MesesGarantiaAlug,omitempty"`            //	Number(2)	Número de meses de garantia do aluguel. Valor default é '0'.
	MesesGarantiaEnc             *int                         `json:"MesesGarantiaEnc,omitempty"`             //	Number(2)	Número de meses de garantia dos encargos. Valor default é '0'.
	Matricula                    *string                      `json:"Matricula,omitempty"`                    //	String(20)	Matrícula do imóvel.
	ZonaRegistro                 *string                      `json:"ZonaRegistro,omitempty"`                 //	String(10)	Zona do Cartório de Registro do imóvel.
	ValorVenda                   *types.Float64                     `json:"ValorVenda,omitempty"`                   //	Number(12,2)	Valor de venda do imóvel.
	NomePredio                   *string                      `json:"NomePredio,omitempty"`                   //	String(50)	Nome do prédio do imóvel.
	Latitude                     *types.Float64                     `json:"Latitude,omitempty"`                     //	Number(10,8)	Latitude do imóvel em graus e decimais do grau.
	Longitude                    *types.Float64                     `json:"Longitude,omitempty"`                    //	Number(11,8)	Longitude do imóvel em graus e decimais do grau.
	ValorAluguel                 *types.Float64                     `json:"ValorAluguel,omitempty"`                 //	Number(12,2)	Valor de aluguel do imóvel.
	ContratoLoc_Ativo            *string                      `json:"ContratoLoc_Ativo,omitempty"`            //	String(1)	Indica se o contrato de locação está ativo. Valor default é 'N'.
	Imediacao                    *string                      `json:"Imediacao,omitempty"`                    //	String(140)	Descrição das imediações do imóvel.
	DescrCaracteristicas         *string                      `json:"DescrCaracteristicas,omitempty"`         //	String	Descrição das características do imóvel.
	DescrReduzida                *string                      `json:"DescrReduzida,omitempty"`                //	String(30)	Descrição reduzida do imóvel.
	ObsIPTU                      *string                      `json:"ObsIPTU,omitempty"`                      //	String	Observações referentes ao IPTU.
	ObsJuridico                  *string                      `json:"ObsJuridico,omitempty"`                  //	String	Observações referentes a atividades jurídicas.
	ObsAcoes                     *string                      `json:"ObsAcoes,omitempty"`                     //	String	Observações referentes a ações feitas ou a fazer no imóvel.
	ObsSeguros                   *string                      `json:"ObsSeguros,omitempty"`                   //	String	Observações referentes ao seguro.
	ObsCadastro                  *string                      `json:"ObsCadastro,omitempty"`                  //	String	Observações referentes ao cadastro do imóvel.
	ObsCtaPagar                  *string                      `json:"ObsCtaPagar,omitempty"`                  //	String	Observações referentes a contas a pagar.
	ObsDOC                       *string                      `json:"ObsDOC,omitempty"`                       //	String	Observações referentes ao cálculo do DOC.
	ObsTaxasCond                 *string                      `json:"ObsTaxasCond,omitempty"`                 //	String	Observações referentes às taxas de condomínio.
	LoginAdmCondom               *string                      `json:"LoginAdmCondom,omitempty"`               //	String(20)	Login de acesso as administradoras de condomínio.
	SenhaAdmCondom               *string                      `json:"SenhaAdmCondom,omitempty"`               //	String(32)	Senha de acesso as administradoras de condomínio. OBSERVAÇÃO: Para fins de segurança, a senha informada neste campo vem criptografada e deve ser um tratamento específico. Ao invés de ser comparada diretamente com a senha digitada pelo usuário, a senha digitada deve ser convertida para maiúsculo e então criptografada em MD5. O valor obtido em MD5 é que deve ser usada na comparação. Exemplo em pseudo-linguagem:
	ObsOutras                    *string                      `json:"ObsOutras,omitempty"`                    //	String	Observações gerais.
	ObsInternet                  *string                      `json:"ObsInternet,omitempty"`                  //	String	Observações que devem ser enviadas para o site na internet.
	ValorCondominio              *types.Float64                     `json:"ValorCondominio,omitempty"`              //	Number(12,2)	Valor mensal do condomínio do imóvel. Valor default é '0'.
	ValorIPTU                    *types.Float64                     `json:"ValorIPTU,omitempty"`                    //	Number(12,2)	Valor mensal de IPTU do imóvel. Valor default é '0'.
	NroInscricaoIPTU             *int                         `json:"NroInscricaoIPTU,omitempty"`             //	Number(17)	Número de inscrição do IPTU.
	InformativoDOC               *string                      `json:"InformativoDOC,omitempty"`               //	String	Texto que deve constar na área do informativo do DOC.
	InstrucaoDOC                 *string                      `json:"InstrucaoDOC,omitempty"`                 //	String	Texto que deve constar na área de instruções do DOC.
	IncideIRFTxAdm               *string                      `json:"IncideIRFTxAdm,omitempty"`               //	String(1)	Indica se incide imposto de renda sobre a taxa de administração. Valor default é 'S'.
	FormaCalcPagto               *string                      `json:"FormaCalcPagto,omitempty"`               //	String(1)	Indica a forma de cálculo para o pagamento ao proprietário. Valor default é 'P'.
	TaxaIntermediacao            *types.Float64                     `json:"TaxaIntermediacao,omitempty"`            //	Number(5,2)	Percentual da taxa de intermediação.
	IncidenciaTaxaAdm            *string                      `json:"IncidenciaTaxaAdm,omitempty"`            //	String(1)	Incidência da taxa de administração. Valor default é 'T'.
	TaxaAdm                      *types.Float64                     `json:"TaxaAdm,omitempty"`                      //	Number(5,2)	Taxa de administração do imóvel em forma de um percentual sobre o aluguel. Se for um valor fixo em Reais então informá-lo no campo 'ValorTaxaAdm' mas apenas um deles deve ser informado.
	ValorTaxaAdm                 *types.Float64                     `json:"ValorTaxaAdm,omitempty"`                 //	Number(11,2)	Taxa de administração do imóvel em forma de um valor fixo em Reais. Se for um percentual sobre o aluguel então informá-lo no campo 'TaxaAdm' mas apenas um deles deve ser informado.
	IncidenciaValorMinimoTaxaAdm *string                      `json:"IncidenciaValorMinimoTaxaAdm,omitempty"` //	String(1)	Indicação de cláusula de valor mínimo de taxa de administração.
	ValorMinimoTaxaAdm           *types.Float64                     `json:"ValorMinimoTaxaAdm,omitempty"`           //	Number(15,2)	Valor mínimo de taxa de administração quando indicado 'Cláusula de valor mínimo de taxa de administração' (IncideValorMinimoTaxaAdm).
	CobrancaAntecipada           *string                      `json:"CobrancaAntecipada,omitempty"`           //	String(1)	Indica se tem desconto de pontualidade quando pago antes do vencimento. Valor default é 'N'.
	RamalAgua                    *string                      `json:"RamalAgua,omitempty"`                    //	String(15)	Identificação do ramal/registro de água.
	CodAgencia                   *int                         `json:"CodAgencia,omitempty"`                   //	Number(5)	Código da agência/loja de captação do imóvel.
	OrigemCaptacao               *string                      `json:"OrigemCaptacao,omitempty"`               //	String(2)	Descrição da origem da captação do imóvel. Valor default é 'O'.
	CodAgenciador                *int                         `json:"CodAgenciador,omitempty"`                //	Number(7)	Código do agenciador de captação do imóvel.
	CodFornecCond                *int                         `json:"CodFornecCond,omitempty"`                //	Number(7)	Código de fornecedor da administradora de condomínio.
	GrupoAnalise                 *int                         `json:"GrupoAnalise,omitempty"`                 //	Number(2)	Grupo de análise. Valor default é '0'.
	GaranteAluguel               *string                      `json:"GaranteAluguel,omitempty"`               //	String(1)	Indica se tem garantia de aluguel. Valor default é 'N'.
	GaranteEncargos              *string                      `json:"GaranteEncargos,omitempty"`              //	String(1)	Indica se tem garantia de encargos. Valor default é 'N'.
	CobraDiferencaCond           *string                      `json:"CobraDiferencaCond,omitempty"`           //	String(1)	Indica se cobra diferença de condomínio. Valor default é 'N'.
	CodPessoaLocat               *int                         `json:"CodPessoaLocat,omitempty"`               //	Number(7)	Código de pessoa do locatário principal.
	NomeLocat                    *string                      `json:"NomeLocat,omitempty"`                    //	String(100)	Nome do locatário.
	CodIntegracaoSist            *string                      `json:"CodIntegracaoSist,omitempty"`            //	String(40)	Código deste imóvel no sistema integrado/migrado.
	CodImovelAgrupar             *int                         `json:"CodImovelAgrupar,omitempty"`             //	Number(8)	Código do imóvel a agrupar após inclusão.
	Caracteristicas              *[]ActionInputCaracteristica `json:"Caracteristicas,omitempty"`              //
	VencMesComp                  *string                      `json:"VencMesComp,omitempty"`                  // Indica vencimento no mês de competência. Valor default é 'N'.
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
