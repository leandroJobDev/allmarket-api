package usecase

import (
	"allmarket/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiService struct {
	apiKey string
}

func NewGeminiService() *GeminiService {
	return &GeminiService{
		apiKey: os.Getenv("GEMINI_API_KEY"),
	}
}

func (s *GeminiService) CategorizarELimparItens(itens []entity.Item) ([]entity.Item, error) {
	if s.apiKey == "" || s.apiKey == "sua_chave_aqui" {
		return itens, fmt.Errorf("GEMINI_API_KEY não configurada")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return itens, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-flash-latest")

	var nomesOriginais []string
	for _, item := range itens {
		nomesOriginais = append(nomesOriginais, item.Nome)
	}

	prompt := fmt.Sprintf(`
Você é um assistente especializado em organizar listas de compras. 
Abaixo está uma lista de nomes de produtos extraídos de uma nota fiscal.
Para cada item:
1. Limpe o nome: Remova códigos, pesos, marcas irrelevantes ou abreviações técnicas.
2. Categorize: Atribua uma categoria (ex: Alimentos, Bebidas, Higiene, Limpeza, Hortifruti, Carnes, Padaria, Outros).

Retorne EXATAMENTE um JSON:
[
  {"original": "FEIJAO PRETO T1 KICALDO 1KG", "limpo": "Feijão Preto", "categoria": "Alimentos"}
]

Lista:
%s
`, strings.Join(nomesOriginais, "\n"))

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return itens, err
	}

	if len(resp.Candidates) == 0 {
		return itens, fmt.Errorf("nenhuma resposta da IA")
	}

	var aiResponse []struct {
		Original  string `json:"original"`
		Limpo     string `json:"limpo"`
		Categoria string `json:"categoria"`
	}

	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		responseText += fmt.Sprintf("%v", part)
	}

	cleanJSON := formatJSONResponse(responseText)
	if err := json.Unmarshal([]byte(cleanJSON), &aiResponse); err != nil {
		return itens, fmt.Errorf("erro JSON: %v", err)
	}

	for i := range itens {
		for _, aiItem := range aiResponse {
			if aiItem.Original == itens[i].Nome {
				itens[i].Nome = aiItem.Limpo
				itens[i].Categoria = aiItem.Categoria
				break
			}
		}
	}

	return itens, nil
}

func (s *GeminiService) ProcessarEstabelecimento(est entity.Estabelecimento) (entity.Estabelecimento, error) {
	if s.apiKey == "" || s.apiKey == "sua_chave_aqui" {
		return est, fmt.Errorf("GEMINI_API_KEY não configurada")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return est, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-flash-latest")

	cnpjRaiz := ""
	digits := ""
	for _, r := range est.CNPJ {
		if r >= '0' && r <= '9' {
			digits += string(r)
		}
	}
	if len(digits) >= 8 {
		cnpjRaiz = digits[:8]
	}

	prompt := fmt.Sprintf(`
Com base no CNPJ Raiz e Razão Social abaixo, identifique o "Nome Fantasia" (nome popular/comercial) do estabelecimento.
Razão Social: %s
CNPJ Completo: %s
CNPJ Raiz: %s

Retorne EXATAMENTE um JSON:
{"nome_fantasia": "Nome Popular do Local"}
`, est.Nome, est.CNPJ, cnpjRaiz)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return est, err
	}

	if len(resp.Candidates) == 0 {
		return est, fmt.Errorf("nenhuma resposta")
	}

	var aiResult struct {
		NomeFantasia string `json:"nome_fantasia"`
	}

	responseText := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		responseText += fmt.Sprintf("%v", part)
	}

	cleanJSON := formatJSONResponse(responseText)
	if err := json.Unmarshal([]byte(cleanJSON), &aiResult); err != nil {
		fmt.Printf("❌ Erro ao decodificar JSON do Gemini: %v | JSON Bruto: [%s]\n", err, cleanJSON)
		return est, err
	}

	if aiResult.NomeFantasia != "" {
		est.NomeFantasia = aiResult.NomeFantasia
		fmt.Printf("✅ Nome Fantasia encontrado: %s\n", est.NomeFantasia)
	} else {
		fmt.Printf("⚠️ Gemini não retornou nome_fantasia para: %s\n", est.Nome)
	}

	return est, nil
}

func formatJSONResponse(text string) string {
	text = strings.TrimSpace(text)
	
	// Remove blocos de código markdown se existirem
	if start := strings.Index(text, "```json"); start != -1 {
		text = text[start+7:]
	} else if start := strings.Index(text, "```"); start != -1 {
		text = text[start+3:]
	}
	
	if end := strings.LastIndex(text, "```"); end != -1 {
		text = text[:end]
	}

	// Tenta encontrar o primeiro { e o último } para garantir que temos apenas o objeto JSON
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end != -1 && end > start {
		text = text[start : end+1]
	}

	return strings.TrimSpace(text)
}
