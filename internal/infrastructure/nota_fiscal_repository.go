package infrastructure

import (
	"allmarket/internal/entity"
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NotaFiscalRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewNotaFiscalRepository(uri string) (*NotaFiscalRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no driver: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao dar ping no Atlas: %w", err)
	}

	db := client.Database("allmarket")
	collection := db.Collection("notas")

	return &NotaFiscalRepository{
		client:     client,
		collection: collection,
	}, nil
}

func (r *NotaFiscalRepository) Salvar(nota entity.NotaFiscal) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"chave": nota.Chave}
	opts := options.Replace().SetUpsert(true)

	_, err := r.collection.ReplaceOne(ctx, filter, nota, opts)
	if err != nil {
		return err
	}
	return nil
}

func (r *NotaFiscalRepository) BuscarPorChave(chave string) (entity.NotaFiscal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var nota entity.NotaFiscal
	filter := bson.M{"chave": strings.TrimSpace(chave)}

	err := r.collection.FindOne(ctx, filter).Decode(&nota)
	return nota, err
}

func (r *NotaFiscalRepository) ListarPorEmail(email string) ([]entity.NotaFiscal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"usuario_email": strings.ToLower(strings.TrimSpace(email))}
	opts := options.Find().SetSort(bson.D{{Key: "data_emissao", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notas []entity.NotaFiscal

	for cursor.Next(ctx) {
		var n entity.NotaFiscal
		if err := cursor.Decode(&n); err != nil {
			continue
		}
		notas = append(notas, n)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	if notas == nil {
		return []entity.NotaFiscal{}, nil
	}

	return notas, nil
}

func (r *NotaFiscalRepository) BuscarTodas() ([]entity.NotaFiscal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	notas := []entity.NotaFiscal{}
	if err = cursor.All(ctx, &notas); err != nil {
		return nil, err
	}

	return notas, nil
}
func (r *NotaFiscalRepository) DeletarPorChaveEEmail(chave string, email string) error {
    ctx := context.TODO()
    filter := bson.M{"chave": chave, "usuario_email": email}
    _, err := r.collection.DeleteOne(ctx, filter)
    return err
}