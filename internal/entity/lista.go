package entity

import "time"

type ItemLista struct {
	Nome     string `json:"nome" firestore:"nome"`
	Comprado bool   `json:"comprado" firestore:"comprado"`
}

type Lista struct {
	ID           string      `json:"id" firestore:"id,omitempty"`
	UsuarioEmail string      `json:"usuario_email" firestore:"usuario_email"`
	Nome         string      `json:"nome" firestore:"nome"`
	Itens        []ItemLista `json:"itens" firestore:"itens"`
	DataCriacao  string      `json:"data_criacao" firestore:"data_criacao"`
	Ativa           bool        `json:"ativa" firestore:"ativa"`
	CompartilhadaCom []string    `json:"compartilhada_com" firestore:"compartilhada_com"`
}

func NewLista(email, nome string, itens []ItemLista, ativa bool) Lista {
	return Lista{
		UsuarioEmail: email,
		Nome:         nome,
		Itens:        itens,
		DataCriacao:  time.Now().Format(time.RFC3339),
		Ativa:        ativa,
	}
}
