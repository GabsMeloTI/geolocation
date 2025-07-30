package plate

type Multa struct {
	NumeroAutoInfracao               string `json:"numeroautoinfracao"`
	DataDaInfracao                   string `json:"datadainfracao"`
	Exigibilidade                    string `json:"exigibilidade"`
	Infracao                         string `json:"infracao"`
	Orgao                            string `json:"orgao"`
	ConsultaDetalheExisteErro        string `json:"consultadetalhe_existe_erro"`
	ConsultaDetalheMensagem          string `json:"consultadetalhe_mensagem"`
	DetalheDataCadastramento         string `json:"detalhe_cadastramento_infracao"`
	DetalheCodInfracao               string `json:"detalhe_cod_infracao"`
	DetalheCodMunEmplacamento        string `json:"detalhe_cod_mun_emplacamento"`
	DetalheCodMunInfracao            string `json:"detalhe_cod_mun_infracao"`
	DetalheDataEmissaoPenalidade     string `json:"detalhe_dt_emissao_penalidade"`
	DetalheDataInfracao              string `json:"detalhe_dt_infracao"`
	DetalheDataNotificacao           string `json:"detalhe_dt_notificacao_infracao"`
	DetalheHoraInfracao              string `json:"detalhe_hr_infracao"`
	DetalheLimitePermitido           string `json:"detalhe_limite_permitido"`
	DetalheLocalInfracao             string `json:"detalhe_local_infracao"`
	DetalheAMRCAModelo               string `json:"detalhe_amrcamodelo"`
	DetalheMedicaoConsiderada        string `json:"detalhe_medicao_considerada"`
	DetalheMedicaoReal               string `json:"detalhe_medicao_real"`
	DetalheNumAutoInfracao           string `json:"detalhe_num_auto_infracao"`
	DetalheOrgaoAutuador             string `json:"detalhe_orgao_autuador"`
	DetalhePlaca                     string `json:"detalhe_placa"`
	DetalheTipoAutoInfracao          string `json:"detalhe_tipo_auto_infracao"`
	DetalheUFJurisdicaoVeiculo       string `json:"detalhe_uf_jurisdicao_veiculo"`
	DetalheUFOrgaoAutuador           string `json:"detalhe_uf_orgao_autuador"`
	DetalheUFPlaca                   string `json:"detalhe_uf_placa"`
	DetalheUnidadeMedida             string `json:"detalhe_unidade_medida"`
	DetalheValorInfracao             string `json:"detalhe_valor_infracao"`
	DadosSuspensaoAceiteUFJurisdicao string `json:"dadosdasuspensao_aceite_uf_jurisdicao"`
	DadosSuspensaoDataRegistro       string `json:"dadosdasuspensao_data_registro"`
	DadosSuspensaoOrigem             string `json:"dadosdasuspensao_origem"`
	DadosSuspensaoTipo               string `json:"dadosdasuspensao_tipo"`
	CNHCondutor                      string `json:"dadosinfrator_cnh_condutor"`
	CNHInfrator                      string `json:"dadosinfrator_cnh_infrator"`
	DtPagamento                      string `json:"dadosdopagamento_dt_pagamento"`
	DtRegistroPagamento              string `json:"dadosdopagamento_dt_do_registro_do_pgmto"`
	UfPagamento                      string `json:"dadosdopagamento_uf_pagamento"`
	ValorPago                        string `json:"dadosdopagamento_valor_pago"`
	DadosPagamento                   string `json:"dadosdopagamento_dados_pgmto"`
}

type FullAPIResponse struct {
	Error   bool     `json:"error"`
	Message string   `json:"message"`
	Data    Response `json:"response"`
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
