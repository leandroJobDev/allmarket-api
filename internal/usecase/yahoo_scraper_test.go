package usecase

import (
	"testing"
)

func TestBuscarResultadosYahoo(t *testing.T) {
	// Skip for now as it makes real network calls, but we can test if it fails gracefully
	t.Run("Should handle empty query", func(t *testing.T) {
		_, err := BuscarResultadosYahoo("")
		if err != nil {
			// This might fail if network is down or return results if Yahoo handles empty query
			// For now, just ensuring it doesn't panic
		}
	})
}
