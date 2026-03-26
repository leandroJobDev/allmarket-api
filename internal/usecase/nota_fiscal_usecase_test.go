package usecase

import (
	"strings"
	"testing"
)

func TestProcessarURL(t *testing.T) {
	t.Run("Deve retornar erro se não for um link (sem prefixo http)", func(t *testing.T) {
		input := "<html><body>Nota</body></html>"
		_, err := ProcessarURL(input)

		if err == nil {
			t.Error("Deveria ter retornado erro para HTML bruto, pois agora a validação exige prefixo http")
		}
		expectedErr := "por favor, insira um link válido"
		if err != nil && !strings.Contains(err.Error(), expectedErr) {
			t.Errorf("Mensagem de erro esperada: '%s', obtida: '%v'", expectedErr, err)
		}
	})

	t.Run("Deve reconhecer URL de São Paulo", func(t *testing.T) {
		url := "https://www.nfce.fazenda.sp.gov.br/consulta?p=112233"
		_, err := ProcessarURL(url)
		
		if err != nil && strings.Contains(err.Error(), "ainda não é suportada") {
			t.Errorf("Deveria reconhecer SP, mas retornou: %v", err)
		}
	})

	t.Run("Deve reconhecer URL da Paraíba", func(t *testing.T) {
		url := "https://www.sefaz.pb.gov.br/nfce?p=998877"
		_, err := ProcessarURL(url)
		
		if err != nil && strings.Contains(err.Error(), "ainda não é suportada") {
			t.Errorf("Deveria reconhecer PB, mas retornou: %v", err)
		}
	})

	t.Run("Deve retornar erro para SEFAZ de estado não mapeado", func(t *testing.T) {
		url := "https://sefaz.rj.gov.br/consulta?p=000"
		_, err := ProcessarURL(url)
		
		if err == nil {
			t.Error("Deveria retornar erro para um estado (RJ) que não está no switch")
		}
		if err != nil && !strings.Contains(err.Error(), "ainda não é suportada") {
			t.Errorf("Erro esperado de suporte, obtido: %v", err)
		}
	})

	t.Run("Deve processar corretamente URL com espaços em branco", func(t *testing.T) {
		url := "   https://fazenda.sp.gov.br/p=123   "
		_, err := ProcessarURL(url)
		
		if err != nil && strings.Contains(err.Error(), "link válido") {
			t.Error("Falhou ao fazer o TrimSpace na URL")
		}
	})
}