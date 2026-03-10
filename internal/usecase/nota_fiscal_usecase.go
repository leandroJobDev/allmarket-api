package usecase

import (
	"allmarket/internal/entity"
	"fmt"
	"strings"
)

func ProcessarURL(input string) (entity.NotaFiscal, error) {
	input = strings.TrimSpace(input)

	if !strings.HasPrefix(input, "http") {
		return entity.NotaFiscal{}, fmt.Errorf("por favor, insira um link válido da nota fiscal (URL)")
	}

	// Roteamento por domínio da SEFAZ
	switch {
	case strings.Contains(input, "sef.sc.gov.br"),   // Santa Catarina
		 strings.Contains(input, "sefaz.pe.gov.br"),   // Pernambuco
		 strings.Contains(input, "sefaz.pb.gov.br"),   // Paraíba
		 strings.Contains(input, "fazenda.sp.gov.br"): // São Paulo 
		return ScraperPadraoNacional(input)

	default:
		return entity.NotaFiscal{}, fmt.Errorf("esta URL da SEFAZ ainda não é suportada")
	}
}