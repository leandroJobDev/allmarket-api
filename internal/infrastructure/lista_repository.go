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

type ListaRepository struct {
	client     *firestore.Client
	collection string
}

func NewListaRepository(projectID string) (*ListaRepository, error) {
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

	return &ListaRepository{
		client:     client,
		collection: "tb_listas",
	}, nil
}

func (r *ListaRepository) Salvar(lista entity.Lista) (string, error) {
	ctx := context.Background()

	var docRef *firestore.DocumentRef
	if lista.ID == "" {
		docRef = r.client.Collection(r.collection).NewDoc()
		lista.ID = docRef.ID
	} else {
		docRef = r.client.Collection(r.collection).Doc(lista.ID)
	}

	_, err := docRef.Set(ctx, lista)
	if err != nil {
		return "", err
	}
	return docRef.ID, nil
}

func (r *ListaRepository) ListarPorEmails(emails []string) ([]entity.Lista, error) {
	ctx := context.Background()
	var listas []entity.Lista

	if len(emails) == 0 {
		return []entity.Lista{}, nil
	}

	iter := r.client.Collection(r.collection).
		Where("usuario_email", "in", emails).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var l entity.Lista
		if err := doc.DataTo(&l); err != nil {
			continue
		}
		l.ID = doc.Ref.ID
		listas = append(listas, l)
	}

	if listas == nil {
		return []entity.Lista{}, nil
	}

	return listas, nil
}

func (r *ListaRepository) Deletar(id string) error {
	ctx := context.Background()
	_, err := r.client.Collection(r.collection).Doc(id).Delete(ctx)
	return err
}

func (r *ListaRepository) DeletarTodasPorEmail(email string) error {
	ctx := context.Background()
	iter := r.client.Collection(r.collection).
		Where("usuario_email", "==", strings.ToLower(strings.TrimSpace(email))).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		_, err = doc.Ref.Delete(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
