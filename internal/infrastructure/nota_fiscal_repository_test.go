package infrastructure

import (
	"os"
	"testing"
)

func TestNewNotaFiscalRepository(t *testing.T) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017" 
	}

	t.Run("Deve conectar com sucesso usando URI valida", func(t *testing.T) {
		repo, err := NewNotaFiscalRepository(uri)
		
		if err != nil {
			t.Skip("Pulando teste: MongoDB local não está rodando")
		}

		if repo == nil {
			t.Fatal("O repositório não deveria ser nil")
		}

		if repo.client == nil || repo.collection == nil {
			t.Error("Client ou Collection não foram inicializados")
		}
	})

	t.Run("Deve retornar erro com URI invalida", func(t *testing.T) {
		uriInvalida := "mongodb://usuario:senha@local-inexistente:27017"
		_, err := NewNotaFiscalRepository(uriInvalida)

		if err == nil {
			t.Error("Deveria ter retornado erro para uma URI inacessível")
		}
	})
}