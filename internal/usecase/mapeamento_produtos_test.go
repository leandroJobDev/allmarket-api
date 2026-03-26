package usecase

import (
	"testing"
)

func TestBuscarCategoriaNoDicionario(t *testing.T) {
	s := &GroqService{}
	testes := []struct {
		nome     string
		esperado string
	}{
		{"FEIJÃO PRETO", "ALIMENTOS"},
		{"SHAMPOO ANTICASPA", "HIGIENE"},
		{"SABÃO EM PÓ", "LIMPEZA"},
		{"CERVEJA GELADA", "BEBIDAS"},
		{"PAO FRANCES", "PADARIA"},
		{"PRODUTO DESCONHECIDO", ""},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			resultado := s.buscarCategoriaNoDicionario(tt.nome)
			if resultado != tt.esperado {
				t.Errorf("buscarCategoriaNoDicionario(%s): esperado %s, obteve %s", tt.nome, tt.esperado, resultado)
			}
		})
	}
}

func TestExpandirNome(t *testing.T) {
	s := &GroqService{}
	testes := []struct {
		entrada  string
		esperado string
	}{
		{"FEIJ CARIOCA", "Feijão CARIOCA"},
		{"ARROZ T1", "Arroz T1"},
		{"ABC PROD", "ABC PROD"},
	}

	for _, tt := range testes {
		t.Run(tt.entrada, func(t *testing.T) {
			resultado := s.expandirNome(tt.entrada)
			if resultado != tt.esperado {
				t.Errorf("expandirNome(%s): esperado %s, obteve %s", tt.entrada, tt.esperado, resultado)
			}
		})
	}
}
