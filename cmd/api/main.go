package main

import (
	"allmarket/internal/infrastructure"
	"allmarket/internal/usecase"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RequisicaoProcessar struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}

type RequisicaoLogin struct {
	Token string `json:"token"`
}

func main() {
	_ = godotenv.Load()

	mongoUser := os.Getenv("MONGO_USER")
	mongoPass := os.Getenv("MONGO_PASS")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	clusterAddr := "cluster0.5sz7ony.mongodb.net"
	passEscapada := url.QueryEscape(mongoPass)
	uri := fmt.Sprintf("mongodb+srv://%s:%s@%s/?appName=Cluster0",
		mongoUser, passEscapada, clusterAddr)

	repo, err := infrastructure.NewNotaFiscalRepository(uri)
	if err != nil {
		fmt.Printf("❌ Erro MongoDB: %v\n", err)
		return
	}

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	router.GET("/historico", func(c *gin.Context) {
		email := c.Query("email")
		if email == "" {
			c.JSON(400, gin.H{"error": "E-mail é obrigatório"})
			return
		}

		notas, err := repo.ListarPorEmail(strings.ToLower(email))
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao buscar histórico"})
			return
		}

		c.JSON(200, notas)
	})

	router.DELETE("/historico/:chave", func(c *gin.Context) {
		chave := c.Param("chave")
		email := c.Query("email")

		if chave == "" || email == "" {
			c.JSON(400, gin.H{"error": "Chave e e-mail são obrigatórios"})
			return
		}

		err := repo.DeletarPorChaveEEmail(chave, email)
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao deletar nota"})
			return
		}

		c.JSON(200, gin.H{"message": "Nota removida com sucesso"})
	})

	router.GET("/config", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"google_client_id": os.Getenv("GOOGLE_CLIENT_ID"),
		})
	})

	router.POST("/processar", func(c *gin.Context) {
		var req RequisicaoProcessar
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Dados inválidos"})
			return
		}

		nota, err := usecase.ScraperPadraoNacional(req.URL)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		if nota.Chave == "" {
			c.JSON(422, gin.H{"error": "Não foi possível extrair os dados desta URL. Verifique se é uma nota válida."})
			return
		}

		userEmail := strings.ToLower(req.Email)
		nota.UsuarioEmail = userEmail

		err = repo.Salvar(nota)

		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				c.JSON(409, gin.H{
					"error": "Nota já cadastrada",
					"nota":  nota,
				})
				return
			}
			c.JSON(500, gin.H{"error": "Erro ao salvar no banco"})
			return
		}

		c.JSON(201, nota)
	})

	fmt.Printf("🚀 Servidor rodando na porta %s\n", port)
	router.Run(":" + port)
}