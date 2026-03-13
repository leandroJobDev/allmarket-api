package main

import (
	"allmarket/internal/infrastructure"
	"allmarket/internal/usecase"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type RequisicaoProcessar struct {
	URL   string `json:"url"`
	Email string `json:"email"`
}

func main() {
	_ = godotenv.Load()

	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	repo, err := infrastructure.NewNotaFiscalRepository(projectID)
	if err != nil {
		fmt.Printf("❌ Erro Firestore: %v\n", err)
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
		if chave == "" {
			c.JSON(400, gin.H{"error": "Chave é obrigatória"})
			return
		}

		err := repo.DeletarPorChaveEEmail(chave, "")
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao deletar nota"})
			return
		}

		c.JSON(200, gin.H{"message": "Nota removida com sucesso"})
	})

	router.POST("/processar", func(c *gin.Context) {
		var req RequisicaoProcessar
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Dados inválidos"})
			return
		}

		nota, err := usecase.ProcessarURL(req.URL)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		if nota.Chave == "" {
			c.JSON(422, gin.H{"error": "Não foi possível processar esta nota"})
			return
		}

		nota.UsuarioEmail = strings.ToLower(req.Email)

		err = repo.Salvar(nota)
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao salvar no Firestore"})
			return
		}

		c.JSON(201, nota)
	})

	fmt.Printf("🚀 Servidor rodando na porta %s\n", port)
	router.Run(":" + port)
}
