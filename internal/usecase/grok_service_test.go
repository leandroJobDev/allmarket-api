package usecase

import (
	"testing"
)

func TestLimparNomeBasico(t *testing.T) {
	s := &GroqService{}
	testes := []struct {
		entrada  string
		esperado string
	}{
		{"ARROZ 1KG", "Arroz 1kg"},
		{"MAC ESPAG", "Macarrão Espaguete"},
		{"FEIJ", "Feijão"},
	}

	for _, tt := range testes {
		t.Run(tt.entrada, func(t *testing.T) {
			resultado := s.limparNomeBasico(tt.entrada)
			if resultado != tt.esperado {
				t.Errorf("limparNomeBasico(%s): esperado %s, obteve %s", tt.entrada, tt.esperado, resultado)
			}
		})
	}
}
