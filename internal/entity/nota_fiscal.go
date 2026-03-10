package entity

import (
	"go.mongodb.org/mongo-driver/v2/bson" // Importando da V2
)

type Estabelecimento struct {
	Nome     string `json:"nome" bson:"nome"`
	CNPJ     string `json:"cnpj" bson:"cnpj"`
	Endereco string `json:"endereco" bson:"endereco"`
}

type Item struct {
	Nome          string  `json:"nome" bson:"nome"`
	Codigo        string  `json:"codigo" bson:"codigo"`
	Quantidade    float64 `json:"quantidade" bson:"quantidade"`
	Unidade       string  `json:"unidade" bson:"unidade"`
	PrecoUnitario float64 `json:"preco_unitario" bson:"preco_unitario"`
	PrecoTotal    float64 `json:"preco_total" bson:"preco_total"`
	ValorTotal    float64 `json:"valor_total" bson:"valor_total"`
}

type NotaFiscal struct {
    ID              bson.ObjectID   `bson:"_id,omitempty" json:"id"`
    UsuarioEmail    string          `bson:"usuario_email" json:"usuario_email"`
    Chave           string          `bson:"chave" json:"chave"`
    Numero          string          `bson:"numero" json:"numero"`
    Serie           string          `bson:"serie" json:"serie"`
    DataEmissao     string          `bson:"data_emissao" json:"data_emissao"`
    Estabelecimento Estabelecimento `bson:"estabelecimento" json:"estabelecimento"`
    Itens           []Item          `bson:"itens" json:"itens"`
    ValorTotal      float64         `bson:"valor_total" json:"valor_total"`
}

func (n NotaFiscal) CalcularTotalDosItens() float64 {
	var total float64
	for _, item := range n.Itens {
		total += item.PrecoTotal
	}
	return total
}