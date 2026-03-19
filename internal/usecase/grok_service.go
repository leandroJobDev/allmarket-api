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

// ItemProcessamento auxilia na organização do fluxo Regras > Cache > IA
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
		// 1. Regras de Negócio Estáticas
		if categoria := s.buscarEmRegrasSimples(item.Nome); categoria != "" {
			item.Categoria = categoria
			// Limpeza básica se resolvido por regra
			item.Nome = s.limparNomeBasico(item.Nome)
			resultados[i] = item
			continue
		}

		// 2. Placeholder para Dicionário/Cache (Buscando itens já aprendidos)
		// TODO: Implementar busca no DB/Redis aqui para reduzir chamadas de IA
		/*
			if itemCache, err := s.buscarNoCache(item.Nome); err == nil {
				resultados[i] = itemCache
				continue
			}
		*/

		// 3. Fallback para IA
		indicesParaIA = append(indicesParaIA, i)
		itensParaIA = append(itensParaIA, item)
	}

	// Só chama a IA se houver itens não resolvidos
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

func (s *GroqService) buscarEmRegrasSimples(nome string) string {
	// Delegando para as tabelas de mapeamento em mapeamento_produtos.go
	return s.buscarCategoriaNoDicionario(nome)
}

func (s *GroqService) limparNomeBasico(nome string) string {
	// Tenta expandir siglas primeiro antes de aplicar a formatação Title
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
		resultados[i] = itens[i] // Default para o original
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