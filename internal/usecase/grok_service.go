package usecase

import (
	"allmarket/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type GroqService struct {
	client *openai.Client
}

func NewGroqService() *GroqService {
	config := openai.DefaultConfig(os.Getenv("GROQ_API_KEY"))
	config.BaseURL = "https://api.groq.com/openai/v1"
	return &GroqService{
		client: openai.NewClientWithConfig(config),
	}
}

type ItemProcessamento struct {
	Original  string
	Resolvido bool
	Item      entity.Item
}

func (s *GroqService) CategorizarELimparItens(itens []entity.Item) ([]entity.Item, error) {
	if len(itens) == 0 {
		return itens, nil
	}

	resultados := make([]entity.Item, len(itens))
	var itensParaIA []entity.Item
	indicesParaIA := make([]int, 0)

	for i, item := range itens {
		if categoria := s.buscarEmRegrasSimples(item.Nome); categoria != "" {
			item.Categoria = categoria
			item.Nome = s.limparNomeBasico(item.Nome)
			resultados[i] = item
			continue
		}

		indicesParaIA = append(indicesParaIA, i)
		itensParaIA = append(itensParaIA, item)
	}

	if len(itensParaIA) > 0 {
		itensProcessadosIA, err := s.processarComIA(itensParaIA)
		if err == nil {
			for idx, itemIA := range itensProcessadosIA {
				originalIdx := indicesParaIA[idx]
				resultados[originalIdx] = itemIA
			}
		} else {
			fmt.Printf("⚠️ Falha na IA: %v. Mantendo itens originais.\n", err)
			for _, originalIdx := range indicesParaIA {
				resultados[originalIdx] = itens[originalIdx]
			}
		}
	}

	return resultados, nil
}

func (s *GroqService) IdentificarEstabelecimento(est entity.Estabelecimento) (entity.Estabelecimento, error) {
	if est.NomeFantasia != "" {
		return est, nil
	}

	radical := GerarCNPJMatriz(est.CNPJ)
	query := fmt.Sprintf("CNPJ %s NOME FANTASIA", radical)
	snippets, _ := BuscarResultadosYahoo(query)
	searchContext := strings.Join(snippets, "\n---\n")

	prompt := fmt.Sprintf(`Atue como um especialista em mercado varejista brasileiro. 
Dado a Razão Social, o Radical do CNPJ, o ENDEREÇO e RESULTADOS DE BUSCA NA WEB, retorne o NOME FANTASIA (marca) mais conhecido.

RAZÃO SOCIAL: %s
RADICAL CNPJ: %s
ENDEREÇO: %s

RESULTADOS DE BUSCA (Contexto):
%s

Regras:
1. Identifique a marca comercial (ex: "Assaí Atacadista", "Pão de Açúcar", "Extra").
2. Ignore nomes de holdings genéricos (ex: "Sendas Distribuidora" deve ser "Assaí Atacadista").
3. Priorize os nomes comerciais encontrados nos resultados de busca.
4. Retorne APENAS um JSON no formato: {"nome_fantasia": "..."}`, est.Nome, radical, est.Endereco, searchContext)

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "llama-3.3-70b-versatile",
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
			Temperature:    0.0,
		},
	)

	if err != nil {
		return est, err
	}

	var aiOutput struct {
		NomeFantasia string `json:"nome_fantasia"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &aiOutput); err != nil {
		return est, err
	}

	if aiOutput.NomeFantasia != "" {
		est.NomeFantasia = strings.TrimSpace(strings.ToUpper(aiOutput.NomeFantasia))
	}

	return est, nil
}

func (s *GroqService) buscarEmRegrasSimples(nome string) string {
	return s.buscarCategoriaNoDicionario(nome)
}

func (s *GroqService) limparNomeBasico(nome string) string {
	expandido := s.expandirNome(nome)
	return strings.Title(strings.ToLower(strings.TrimSpace(expandido)))
}

func (s *GroqService) processarComIA(itens []entity.Item) ([]entity.Item, error) {
	type ItemInput struct {
		Original string  `json:"original"`
		Qtd      float64 `json:"qtd"`
		Uni      string  `json:"uni"`
	}

	var inputs []ItemInput
	for _, item := range itens {
		inputs = append(inputs, ItemInput{
			Original: item.Nome,
			Qtd:      item.Quantidade,
			Uni:      item.Unidade,
		})
	}

	inputJSON, _ := json.Marshal(inputs)

	prompt := fmt.Sprintf(`Atue como um motor de limpeza de dados. 
Limpe nomes de itens de supermercado removendo pesos/volumes (ex: 1kg, 500ml) 
e categorize-os estritamente em: ALIMENTOS, LATICÍNIOS, PADARIA, CARNES E EMBUTIDOS, LIMPEZA, HIGIENE PESSOAL, BEBIDAS, HORTIFRUTI ou OUTROS. 

Retorne APENAS JSON no formato:
{"itens": [{"original": "...", "completo": "...", "qtd": 1.0, "uni": "...", "categoria": "..."}]}

Itens: %s`, string(inputJSON))

	resp, err := s.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "llama-3.3-70b-versatile",
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
			Temperature:    0.0,
		},
	)

	if err != nil {
		return nil, err
	}

	var aiOutput struct {
		Itens []struct {
			Original  string  `json:"original"`
			Completo  string  `json:"completo"`
			Qtd       float64 `json:"qtd"`
			Uni       string  `json:"uni"`
			Categoria string  `json:"categoria"`
		} `json:"itens"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &aiOutput); err != nil {
		return nil, err
	}

	resultados := make([]entity.Item, len(itens))
	for i := range itens {
		resultados[i] = itens[i]
		for _, ai := range aiOutput.Itens {
			if strings.EqualFold(strings.TrimSpace(ai.Original), strings.TrimSpace(itens[i].Nome)) {
				resultados[i].Nome = ai.Completo
				resultados[i].Quantidade = ai.Qtd
				resultados[i].Unidade = strings.ToLower(ai.Uni)
				resultados[i].Categoria = strings.ToUpper(ai.Categoria)
				break
			}
		}
	}

	return resultados, nil
}