package main

import (
	"allmarket/internal/entity"
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

	repoLista, err := infrastructure.NewListaRepository(projectID)
	if err != nil {
		fmt.Printf("❌ Erro Firestore Lista: %v\n", err)
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

	router.GET("/listas", func(c *gin.Context) {
		email := c.Query("email")
		if email == "" {
			c.JSON(400, gin.H{"error": "E-mail é obrigatório"})
			return
		}

		listas, err := repoLista.ListarPorEmail(strings.ToLower(email))
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao buscar listas"})
			return
		}

		c.JSON(200, listas)
	})

	router.POST("/listas", func(c *gin.Context) {
		var lista entity.Lista
		if err := c.ShouldBindJSON(&lista); err != nil {
			c.JSON(400, gin.H{"error": "Dados inválidos"})
			return
		}

		if lista.UsuarioEmail == "" {
			c.JSON(400, gin.H{"error": "E-mail do usuário é obrigatório"})
			return
		}

		id, err := repoLista.Salvar(lista)
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao salvar lista"})
			return
		}

		lista.ID = id
		c.JSON(200, lista)
	})

	router.DELETE("/listas/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(400, gin.H{"error": "ID é obrigatório"})
			return
		}

		err := repoLista.Deletar(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Erro ao deletar lista"})
			return
		}

		c.JSON(200, gin.H{"message": "Lista removida com sucesso"})
	})

	router.POST("/listas/sincronizar", func(c *gin.Context) {
		var req struct {
			Email  string         `json:"email"`
			Listas []entity.Lista `json:"listas"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Dados inválidos"})
			return
		}

		for _, lista := range req.Listas {
			lista.UsuarioEmail = strings.ToLower(req.Email)
			_, _ = repoLista.Salvar(lista)
		}

		c.JSON(200, gin.H{"message": "Sincronização concluída"})
	})

	fmt.Printf("🚀 Servidor rodando na porta %s\n", port)
	router.Run(":" + port)
}
