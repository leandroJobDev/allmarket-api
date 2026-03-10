package entity

import "testing"

func TestCalcularTotalDosItens(t *testing.T) {
	// 1. Cenário: Criamos uma nota com itens fictícios
	nf := NotaFiscal{
		Itens: []Item{
			{Nome: "Item A", PrecoTotal: 10.50},
			{Nome: "Item B", PrecoTotal: 20.00},
			{Nome: "Item C", PrecoTotal: 5.25},
		},
	}

	// 2. Execução: Chamamos o método que você criou
	resultado := nf.CalcularTotalDosItens()

	// 3. Verificação: O total esperado é 35.75
	esperado := 35.75
	if resultado != esperado {
		t.Errorf("Erro no cálculo: esperado %.2f, mas obteve %.2f", esperado, resultado)
	}
}