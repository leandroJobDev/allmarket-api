package infrastructure

import (
	"allmarket/internal/entity"
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type NotaFiscalRepository struct {
	client     *firestore.Client
	collection string
}

func NewNotaFiscalRepository(projectID string) (*NotaFiscalRepository, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no Firestore: %w", err)
	}

	return &NotaFiscalRepository{
		client:     client,
		collection: "tb_notas", // Nome da sua coleção
	}, nil
}

func (r *NotaFiscalRepository) Salvar(nota entity.NotaFiscal) error {
	ctx := context.Background()
	
	// Usamos a chave da nota como ID do documento para facilitar buscas
	// e garantir que a mesma nota não seja salva duas vezes (Upsert)
	_, err := r.client.Collection(r.collection).Doc(nota.Chave).Set(ctx, nota)
	return err
}

func (r *NotaFiscalRepository) BuscarPorChave(chave string) (entity.NotaFiscal, error) {
	ctx := context.Background()
	
	doc, err := r.client.Collection(r.collection).Doc(strings.TrimSpace(chave)).Get(ctx)
	if err != nil {
		return entity.NotaFiscal{}, err
	}

	var nota entity.NotaFiscal
	err = doc.DataTo(&nota)
	return nota, err
}

func (r *NotaFiscalRepository) ListarPorEmail(email string) ([]entity.NotaFiscal, error) {
	ctx := context.Background()
	var notas []entity.NotaFiscal

	// Consulta filtrando pelo email do usuário
	iter := r.client.Collection(r.collection).
		Where("usuario_email", "==", strings.ToLower(strings.TrimSpace(email))).
		OrderBy("data_emissao", firestore.Desc).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var n entity.NotaFiscal
		if err := doc.DataTo(&n); err != nil {
			continue
		}
		notas = append(notas, n)
	}

	return notas, nil
}

func (r *NotaFiscalRepository) DeletarPorChaveEEmail(chave string, email string) error {
	ctx := context.Background()
	// No Firestore, deletamos diretamente pelo ID do documento (chave)
	_, err := r.client.Collection(r.collection).Doc(chave).Delete(ctx)
	return err
}