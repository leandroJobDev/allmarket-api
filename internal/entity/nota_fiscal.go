package entity

type Estabelecimento struct {
	Nome          string `json:"nome" firestore:"nome"`
	CNPJ          string `json:"cnpj" firestore:"cnpj"`
	Endereco      string `json:"endereco" firestore:"endereco"`
	NomeFantasia  string `json:"nome_fantasia" firestore:"nome_fantasia"`
}

type Item struct {
	Nome          string  `json:"nome" firestore:"nome"`
	Codigo        string  `json:"codigo" firestore:"codigo"`
	Quantidade    float64 `json:"quantidade" firestore:"quantidade"`
	Unidade       string  `json:"unidade" firestore:"unidade"`
	PrecoUnitario float64 `json:"preco_unitario" firestore:"preco_unitario"`
	PrecoTotal    float64 `json:"preco_total" firestore:"preco_total"`
	ValorTotal    float64 `json:"valor_total" firestore:"valor_total"`
	Categoria     string  `json:"categoria" firestore:"categoria"`
}

type NotaFiscal struct {
	UsuarioEmail    string          `json:"usuario_email" firestore:"usuario_email"`
	Chave           string          `json:"chave" firestore:"chave"`
	Numero          string          `json:"numero" firestore:"numero"`
	Serie           string          `json:"serie" firestore:"serie"`
	DataEmissao     string          `json:"data_emissao" firestore:"data_emissao"`
	Estabelecimento Estabelecimento `json:"estabelecimento" firestore:"estabelecimento"`
	Itens           []Item          `json:"itens" firestore:"itens"`
	ValorTotal      float64         `json:"valor_total" firestore:"valor_total"`
}

func (n NotaFiscal) CalcularTotalDosItens() float64 {
	var total float64
	for _, item := range n.Itens {
		total += item.PrecoTotal
	}
	return total
}