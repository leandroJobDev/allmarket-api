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

type VinculoRepository struct {
	client     *firestore.Client
	collection string
}

func NewVinculoRepository(projectID string) (*VinculoRepository, error) {
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

	return &VinculoRepository{
		client:     client,
		collection: "tb_vinculos",
	}, nil
}

func (r *VinculoRepository) Vincular(emailA, emailB string) error {
	ctx := context.Background()
	emailA = strings.ToLower(strings.TrimSpace(emailA))
	emailB = strings.ToLower(strings.TrimSpace(emailB))

	if emailA == emailB {
		return fmt.Errorf("não é possível vincular o mesmo e-mail")
	}

	// Ordena para evitar duplicatas A-B e B-A
	id := emailA + "_" + emailB
	if emailA > emailB {
		id = emailB + "_" + emailA
	}

	v := entity.Vinculo{
		EmailA: emailA,
		EmailB: emailB,
	}

	_, err := r.client.Collection(r.collection).Doc(id).Set(ctx, v)
	return err
}

func (r *VinculoRepository) Desvincular(emailA, emailB string) error {
	ctx := context.Background()
	emailA = strings.ToLower(strings.TrimSpace(emailA))
	emailB = strings.ToLower(strings.TrimSpace(emailB))

	id := emailA + "_" + emailB
	if emailA > emailB {
		id = emailB + "_" + emailA
	}

	_, err := r.client.Collection(r.collection).Doc(id).Delete(ctx)
	return err
}

func (r *VinculoRepository) ObterEmailsRelacionados(email string) ([]string, error) {
	ctx := context.Background()
	email = strings.ToLower(strings.TrimSpace(email))
	
	emails := []string{email} // O próprio e-mail sempre está incluso

	// Busca onde é EmailA
	iterA := r.client.Collection(r.collection).Where("email_a", "==", email).Documents(ctx)
	for {
		doc, err := iterA.Next()
		if err == iterator.Done { break }
		if err != nil { return nil, err }
		var v entity.Vinculo
		if err := doc.DataTo(&v); err == nil {
			emails = append(emails, v.EmailB)
		}
	}

	// Busca onde é EmailB
	iterB := r.client.Collection(r.collection).Where("email_b", "==", email).Documents(ctx)
	for {
		doc, err := iterB.Next()
		if err == iterator.Done { break }
		if err != nil { return nil, err }
		var v entity.Vinculo
		if err := doc.DataTo(&v); err == nil {
			emails = append(emails, v.EmailA)
		}
	}

	return emails, nil
}
