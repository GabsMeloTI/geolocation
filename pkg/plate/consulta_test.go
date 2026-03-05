package plate

import (
	"fmt"
	"testing"
)

func TestConsultarPlaca(t *testing.T) {
	// Placa de exemplo que você mandou na solicitação
	placa := "TTF8C62"

	resp, err := ConsultarPlaca(placa)
	if err != nil {
		t.Fatalf("Erro ao consultar placa: %v", err)
	}

	if resp.Error {
		t.Fatalf("Resposta com erro: %s", resp.Message)
	}

	fmt.Printf("Sucesso! Veículo: %s %s - %s\n", resp.Data.Marca, resp.Data.Modelo, resp.Data.Cor)
	fmt.Printf("Ano: %s/%s\n", resp.Data.Ano, resp.Data.AnoModelo)
	fmt.Printf("Multas encontradas: %d\n", len(resp.Data.Multas.Dados))

	for i, multa := range resp.Data.Multas.Dados {
		fmt.Printf("[%d] %s - Valor: %s\n", i+1, multa.Infracao, multa.DetalheValorInfracao)
	}
}

func TestConsultarMultas(t *testing.T) {
	placa := "TTF8C62"

	resp, err := ConsultarMultas(placa)
	if err != nil {
		t.Fatalf("Erro ao consultar multas: %v", err)
	}

	fmt.Printf("Sucesso na consulta de multas isolada! Placa: %s\n", resp.Data.Placa)
	fmt.Printf("Quantidade de ocorrências: %s\n", resp.Data.QuantidadeOcorrencias)

	for _, reg := range resp.Data.Registros {
		fmt.Printf("Infracao: %s Auto: %s\n", reg.Infracao, reg.NumeroAutoInfracao)
	}
}
