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

func (s *GroqService) CategorizarELimparItens(itens []entity.Item) ([]entity.Item, error) {
	if len(itens) == 0 {
		return itens, nil
	}

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

	prompt := fmt.Sprintf(`Atue como um sistema de limpeza e categorização de itens de supermercado.

REGRAS DE OURO (NÃO DESVIE):
1. MANTENHA O NOME ORIGINAL: Não remova marcas (Kicaldo, Solito, Sadia, Do Valle, Marba, etc). Apenas substitua a sigla inicial se ela estiver na tabela abaixo.
2. EXPANSÃO DO 1º TERMO (Substitua apenas o termo abreviado): 
   - "FEIJ" -> "Feijao" | "ARROZ" -> "Arroz" | "P FORMA" -> "Pao De Forma"
   - "P QJ" ou "PAO QJO" -> "Pao De Queijo" | "QJ" ou "MUSS" -> "Queijo"
   - "B LAC" -> "Bebida Lactea" | "GDP" -> "Guardanapo De Papel"
   - "MANT" -> "Manteiga" | "ERV" -> "Ervilha" | "SH" -> "Shampoo"
   - "AP" -> "Aparelho" | "CJ" -> "Conjunto" | "T PAP" -> "Toalha De Papel"
   - "MORT" -> "Mortadela" | "SALS" -> "Salsicha"

3. CATEGORIZAÇÃO OBRIGATÓRIA (PROIBIDO USAR 'OUTROS' PARA ALIMENTOS OU LIMPEZA):
   - ALIMENTOS: Arroz, Feijao, Cafe, Salgadinho, Milho, Ervilha, Palmito, Tomate, Batata.
   - LATICÍNIOS: Queijo, Manteiga, Iogurte, Bebida Lactea, Leite.
   - PADARIA: Pao De Forma, Pao De Queijo, Pao Frances.
   - CARNES E EMBUTIDOS: Bacon, Salsicha, Mortadela.
   - LIMPEZA: Agua Sanitaria, Guardanapo, Toalha Papel, Aromatizador, Limpador, Difusor, Evitamofo.
   - HIGIENE PESSOAL: Shampoo, Desodorante, Aparelho Barbear, Mascara, Condicionador.
   - OUTROS: Apenas ferragens (Gancho, Bucha), Panelas, Baldes, Utilidades domésticas.

4. LIMPEZA: Remova apenas o peso/volume (Ex: "1kg", "500g") do texto final do nome.

Retorne EXATAMENTE no formato JSON abaixo:
{"itens": [{"original": "...", "completo": "...", "qtd": 1.0, "uni": "...", "categoria": "..."}]}

Lista de itens: %s`, string(inputJSON))

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
		return itens, err
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
		fmt.Printf("❌ Erro ao decodificar JSON do Groq: %v\n", err)
		return itens, err
	}

	for i := range itens {
		for _, ai := range aiOutput.Itens {
			if strings.EqualFold(strings.TrimSpace(ai.Original), strings.TrimSpace(itens[i].Nome)) {
				itens[i].Nome = ai.Completo
				itens[i].Quantidade = ai.Qtd
				itens[i].Unidade = strings.ToLower(ai.Uni)
				itens[i].Categoria = strings.ToUpper(ai.Categoria)
				break
			}
		}
	}
	return itens, nil
}