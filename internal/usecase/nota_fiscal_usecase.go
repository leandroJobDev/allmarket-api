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

	var nf entity.NotaFiscal
	var err error

	switch {
	case strings.Contains(input, "sef.sc.gov.br"),
		 strings.Contains(input, "sefaz.pe.gov.br"),
		 strings.Contains(input, "sefaz.pb.gov.br"),
		 strings.Contains(input, "fazenda.sp.gov.br"):
		nf, err = ScraperPadraoNacional(input)

	default:
		return entity.NotaFiscal{}, fmt.Errorf("esta URL da SEFAZ ainda não é suportada")
	}

	if err != nil {
		return nf, err
	}

	gemini := NewGeminiService()
	if nf.Estabelecimento.CNPJ != "" {
		if estEnriquecido, err := gemini.ProcessarEstabelecimento(nf.Estabelecimento); err == nil {
			nf.Estabelecimento = estEnriquecido
		}
	}

	if itensProcessados, err := gemini.CategorizarELimparItens(nf.Itens); err == nil {
		nf.Itens = itensProcessados
	}

	return nf, nil
}