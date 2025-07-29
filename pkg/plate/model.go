package plate

type Multa struct {
	NumeroAutoInfracao       string `json:"numeroautoinfracao"`
	DataDaInfracao           string `json:"datadainfracao"`
	Infracao                 string `json:"infracao"`
	Orgao                    string `json:"orgao"`
	Descricao                string `json:"detalhe_cod_infracao"`
	Local                    string `json:"detalhe_local_infracao"`
	Valor                    string `json:"detalhe_valor_infracao"`
	DetalheDataCadastramento string `json:"detalhe_cadastramento_infracao"`
	DetalheOrgaoAutuador     string `json:"detalhe_orgao_autuador"`
	DetalheUF                string `json:"detalhe_uf_orgao_autuador"`
	DetalhePlaca             string `json:"detalhe_placa"`
	CNHInfrator              string `json:"dadosinfrator_cnh_infrator"`
	ValorPago                string `json:"dadosdopagamento_valor_pago"`
}

type FullAPIResponse struct {
	Error   bool     `json:"error"`
	Message string   `json:"message"`
	Data    Response `json:"response"` // <- era "data", agora é "response"
}

type Response struct {
	MARCA             string `json:"MARCA"`
	MODELO            string `json:"MODELO"`
	SUBMODELO         string `json:"SUBMODELO"`
	VERSAO            string `json:"VERSAO"`
	Ano               string `json:"ano"`
	AnoModelo         string `json:"anoModelo"`
	Chassi            string `json:"chassi"`
	CodigoRetorno     string `json:"codigoRetorno"`
	CodigoSituacao    string `json:"codigoSituacao"`
	Cor               string `json:"cor"`
	Data              string `json:"data"`
	Placa             string `json:"placa"`
	PlacaModeloAntigo string `json:"placa_modelo_antigo"`
	PlacaModeloNovo   string `json:"placa_modelo_novo"`
	PlacaNova         string `json:"placa_nova"`
	Modelo            string `json:"modelo"`
	Marca             string `json:"marca"`
	Municipio         string `json:"municipio"`
	Uf                string `json:"uf"`
	UfPlaca           string `json:"uf_placa"`
	Combustivel       string `json:"combustivel"`
	Potencia          string `json:"potencia"`
	CapacidadeCarga   string `json:"capacidade_carga"`
	Nacionalidade     struct {
		Nacionalidade string `json:"nacionalidade"`
	} `json:"nacionalidade"`
	Linha           string      `json:"linha"`
	Carroceria      interface{} `json:"carroceria"`
	CaixaCambio     string      `json:"caixa_cambio"`
	EixoTraseiroDif string      `json:"eixo_traseiro_dif"`
	TerceiroEixo    string      `json:"terceiro_eixo"`
	AnoModelo1      string      `json:"ano_modelo"`
	TipoVeiculo     struct {
		TipoVeiculo string `json:"tipo_veiculo"`
	} `json:"tipo_veiculo"`
	MarcaModelo struct {
		Modelo   string `json:"modelo"`
		Marca    string `json:"marca"`
		Segmento string `json:"segmento"`
		Versao   string `json:"versao"`
	} `json:"marca_modelo"`
	CorVeiculo struct {
		Cor string `json:"cor"`
	} `json:"cor_veiculo"`
	QuantidadePassageiro string     `json:"quantidade_passageiro"`
	SituacaoChassi       string     `json:"situacao_chassi"`
	Eixos                string     `json:"eixos"`
	TipoMontagem         string     `json:"tipo_montagem"`
	UltimaAtualizacao    string     `json:"ultima_atualizacao"`
	Cilindradas          string     `json:"cilindradas"`
	SituacaoVeiculo      string     `json:"situacao_veiculo"`
	Listamodelo          [][]string `json:"listamodelo"`
	Multas               struct {
		Dados []Multa `json:"dados"`
	} `json:"multas,omitempty"`
}
