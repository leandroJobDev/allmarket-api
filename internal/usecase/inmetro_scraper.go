package usecase

import (
	"allmarket/internal/entity"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func FetchEstabelecimentoByCNPJ(cnpj string) (entity.Estabelecimento, error) {
	cnpjDigits := regexp.MustCompile(`\D`).ReplaceAllString(cnpj, "")
	if len(cnpjDigits) != 14 {
		return entity.Estabelecimento{}, fmt.Errorf("CNPJ inválido: %s", cnpj)
	}

	url := fmt.Sprintf("http://www.inmetro.gov.br/prodcert/empresas/detalhe.asp?codigo_pais=BR&id_empresa=%s", cnpjDigits)

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	res, err := client.Do(req)
	if err != nil {
		return entity.Estabelecimento{}, fmt.Errorf("falha ao conectar ao Inmetro: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return entity.Estabelecimento{}, fmt.Errorf("erro no Inmetro: status %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return entity.Estabelecimento{}, fmt.Errorf("falha ao processar HTML: %v", err)
	}

	est := entity.Estabelecimento{}

	doc.Find("table tr").Each(func(_ int, s *goquery.Selection) {
		label := strings.Join(strings.Fields(s.Find("td.titulo_certificado").Text()), " ")
		value := strings.TrimSpace(s.Find("td.borda_baixo").Text())

		if value == "" {
			return
		}

		switch {
		case strings.Contains(label, "Razão Social"):
			est.Nome = value
		case strings.Contains(label, "Nome Fantasia"):
			est.NomeFantasia = value
		case strings.Contains(label, "CNPJ"):
			est.CNPJ = value
		case strings.Contains(label, "Endereço"):
			est.Endereco = value
		}
	})

	if est.Nome == "" && est.CNPJ == "" {
		return entity.Estabelecimento{}, fmt.Errorf("empresa não encontrada para o CNPJ: %s", cnpj)
	}

	return est, nil
}
