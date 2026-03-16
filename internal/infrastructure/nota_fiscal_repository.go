package infrastructure

import (
	"allmarket/internal/entity"
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type NotaFiscalRepository struct {
	client     *firestore.Client
	collection string
}

func NewNotaFiscalRepository(projectID string) (*NotaFiscalRepository, error) {
	ctx := context.Background()
	
	var client *firestore.Client
	var err error

	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath != "" {
		client, err = firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credsPath))
	} else {
		client, err = firestore.NewClient(ctx, projectID)
	}

	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no Firestore: %w", err)
	}

	return &NotaFiscalRepository{
		client:     client,
		collection: "tb_notas",
	}, nil
}

func (r *NotaFiscalRepository) Salvar(nota entity.NotaFiscal) error {
	ctx := context.Background()

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

func (r *NotaFiscalRepository) ListarPorEmails(emails []string) ([]entity.NotaFiscal, error) {
	ctx := context.Background()
	var notas []entity.NotaFiscal

	if len(emails) == 0 {
		return []entity.NotaFiscal{}, nil
	}

	// Firestore "in" query suporta até 10 elementos. 
	// Para uso doméstico (compartilhar com esposa/filhos) isso é suficiente.
	iter := r.client.Collection(r.collection).
		Where("usuario_email", "in", emails).
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
	_, err := r.client.Collection(r.collection).Doc(chave).Delete(ctx)
	return err
}
