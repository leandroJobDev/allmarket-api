package usecase

import (
	"allmarket/internal/entity"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func ScraperPadraoNacional(urlNota string) (entity.NotaFiscal, error) {
	doc, err := obterDocumento(urlNota)
	if err != nil {
		return entity.NotaFiscal{}, err
	}

	textoCompleto := doc.Text()

	nf := entity.NotaFiscal{
		Chave:           extrairChave(doc, textoCompleto),
		Estabelecimento: extrairEstabelecimento(doc, textoCompleto),
	}

	if doc.Find("det").Length() > 0 {
		return extrairDadosXML(doc, nf), nil
	}

	return extrairDadosHTML(doc, nf, textoCompleto), nil
}

func obterDocumento(input string) (*goquery.Document, error) {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "http") {
		client := &http.Client{Timeout: 30 * time.Second}
		req, _ := http.NewRequest("GET", input, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("falha na conexão: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			return nil, fmt.Errorf("erro na SEFAZ: status %d", res.StatusCode)
		}
		return goquery.NewDocumentFromReader(res.Body)
	}
	return goquery.NewDocumentFromReader(strings.NewReader(input))
}

func extrairChave(doc *goquery.Document, texto string) string {
	chave := doc.Find(".chave").Text()
	if chave == "" {
		chave = doc.Find("#chave, .txtChave").Text()
	}

	if chave == "" {
		re := regexp.MustCompile(`(\d\s*){44}`)
		chave = re.FindString(texto)
	}

	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, chave)
}

func extrairEstabelecimento(doc *goquery.Document, texto string) entity.Estabelecimento {
	nome := doc.Find("emit xNome").Text()
	if nome == "" {
		nome = doc.Find(".txtTopo, #u20, .txtTit").First().Text()
	}

	reCNPJ := regexp.MustCompile(`\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}`)
	cnpj := doc.Find("emit CNPJ").Text()
	if cnpj == "" {
		cnpj = reCNPJ.FindString(texto)
	}

	return entity.Estabelecimento{
		Nome:     strings.TrimSpace(nome),
		CNPJ:     cnpj,
		Endereco: extrairEndereco(doc),
	}
}

func extrairEndereco(doc *goquery.Document) string {
	if rua := doc.Find("enderEmit xLgr").Text(); rua != "" {
		return fmt.Sprintf("%s, %s - %s", rua, doc.Find("enderEmit nro").Text(), doc.Find("enderEmit xMun").Text())
	}

	var partes []string
	doc.Find(".text, .txtEndereco").Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" && !strings.Contains(t, "CNPJ") && (strings.Contains(t, ",") || strings.Contains(t, "SP")) {
			partes = append(partes, t)
		}
	})
	
	if len(partes) == 0 { return "Endereço não identificado" }
	return strings.Join(strings.Fields(strings.Join(partes, " ")), " ")
}

func extrairDadosXML(doc *goquery.Document, nf entity.NotaFiscal) entity.NotaFiscal {
	nf.Numero = doc.Find("ide nNF").Text()
	nf.Serie = doc.Find("ide serie").Text()
	nf.DataEmissao = normalizarData(doc.Find("ide dhEmi").Text())
	nf.ValorTotal = extrairNumero(doc.Find("total vNF").Text())

	doc.Find("det").Each(func(_ int, s *goquery.Selection) {
		nf.Itens = append(nf.Itens, entity.Item{
			Nome:          strings.Join(strings.Fields(s.Find("xProd").Text()), " "),
			Codigo:        s.Find("cProd").Text(),
			Quantidade:    extrairNumero(s.Find("qCom").Text()),
			Unidade:       s.Find("uCom").Text(),
			PrecoUnitario: extrairNumero(s.Find("vUnCom").Text()),
			PrecoTotal:    extrairNumero(s.Find("vProd").Text()),
		})
	})
	return nf
}

func extrairDadosHTML(doc *goquery.Document, nf entity.NotaFiscal, texto string) entity.NotaFiscal {
	nf.Numero = regexBusca(texto, `(?i)Número:\s*(\d+)`)
	nf.Serie = regexBusca(texto, `(?i)Série:\s*(\d+)`)
	nf.DataEmissao = normalizarData(regexBusca(texto, `(?i)Emissão:\s*(\d{2}/\d{2}/\d{4}\s*\d{2}:\d{2}:\d{2})`))
	
	nf.ValorTotal = extrairNumero(regexBusca(texto, `(?i)Valor\s*a\s*pagar\s*R\$\s*([0-9.,]+)`))
	if nf.ValorTotal == 0 {
		nf.ValorTotal = extrairNumero(doc.Find(".valor, .totalNFe, .txtMax").Last().Text())
	}

	doc.Find("#tabResult tr, .table tr").Each(func(_ int, s *goquery.Selection) {
		nome := strings.TrimSpace(s.Find(".txtTit").First().Text())
		if nome == "" || strings.Contains(nome, "Vl. Total") { return }

		nf.Itens = append(nf.Itens, entity.Item{
			Nome:          strings.Join(strings.Fields(nome), " "),
			Codigo:        regexBusca(s.Find(".RCod, .txtCodigo").Text(), `\d+`),
			Quantidade:    extrairNumero(s.Find(".Rqtd, .txtQtde").Text()),
			Unidade:       strings.TrimSpace(strings.Replace(s.Find(".RUN, .txtUnidade").Text(), "UN:", "", 1)),
			PrecoUnitario: extrairNumero(s.Find(".RvlUnit, .txtValor").Text()),
			PrecoTotal:    extrairNumero(s.Find(".valor, .txtValorTotal").Text()),
		})
	})
	return nf
}

func regexBusca(texto, padrao string) string {
	re := regexp.MustCompile(padrao)
	m := re.FindStringSubmatch(texto)
	if len(m) > 1 { return m[1] }
	return ""
}

func normalizarData(dataBruta string) string {
	dataBruta = strings.TrimSpace(dataBruta)
	if strings.Contains(dataBruta, "-") && strings.Contains(dataBruta, "T") {
		t, _ := time.Parse(time.RFC3339, dataBruta)
		return t.Format("02/01/2006 15:04:05")
	}
	return dataBruta
}

func extrairNumero(texto string) float64 {
	if texto == "" { return 0 }

	limpo := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == ',' || r == '.' { return r }
		return -1
	}, texto)

	if strings.Contains(limpo, ",") {
		limpo = strings.ReplaceAll(limpo, ".", "") 
		limpo = strings.Replace(limpo, ",", ".", 1)
	}

	v, _ := strconv.ParseFloat(limpo, 64)
	return v
}