package usecase

import (
	"testing"
)

func TestFetchEstabelecimentoByCNPJ(t *testing.T) {
	cnpj := "00.000.000/0001-91"
	est, err := FetchEstabelecimentoByCNPJ(cnpj)
	if err != nil {
		t.Fatalf("Erro ao buscar estabelecimento: %v", err)
	}

	if est.Nome == "" {
		t.Error("Razão Social (Nome) não deveria estar vazia")
	}

	if est.NomeFantasia == "" {
		t.Error("Nome Fantasia não deveria estar vazio")
	}

	if est.CNPJ == "" {
		t.Error("CNPJ não deveria estar vazio")
	}

	if est.Endereco == "" {
		t.Error("Endereço não deveria estar vazio")
	}

	t.Logf("Empresa encontrada: %+v", est)
}
