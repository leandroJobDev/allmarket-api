package usecase

import "testing"

func TestExtrairNumero(t *testing.T) {
	testes := []struct {
		nome     string
		entrada  string
		esperado float64
	}{
		{"Padrão BR", "13,90", 13.90},
		{"Padrão PE (sem ponto)", "139000", 13.90},
		{"Com símbolo R$", "R$ 45,50", 45.50},
		{"Quantidade Inteira", "2.0000", 2.0},
		{"Valor Vazio", "", 0.0},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			resultado := extrairNumero(tt.entrada)
			if resultado != tt.esperado {
				t.Errorf("ExtrairNumero(%s): esperado %.2f, obteve %.2f", tt.entrada, tt.esperado, resultado)
			}
		})
	}
}