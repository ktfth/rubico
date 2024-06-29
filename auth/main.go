package main

import (
	"context"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"github.com/ServiceWeaver/weaver"
	"go.mongodb.org/mongo-driver/mongo"

	"auth/config"
	"auth/routes"
)

// api is the main application component.
type api struct {
	weaver.Implements[weaver.Main]
	routes.AuthRoutes
	mongodbClient *mongo.Client
}

// Start is called by Service Weaver to start the API component.
func (a *api) Start(ctx context.Context) error {
	client, err := config.ConnectDB(ctx)
	if err != nil {
		return err
	}
	a.mongodbClient = client
	log.Println("Connected to MongoDB")

	// Inicializa as rotas
	a.AuthRoutes.DB = a.mongodbClient // Passa o cliente MongoDB para AuthRoutes

	// Wrapper para RegisterLogin
	registerLoginHandler := func(w http.ResponseWriter, req *http.Request) {
		a.AuthRoutes.RegisterLogin(ctx, w, req)
	}

	// Wrapper para Verify
	verifyHandler := func(w http.ResponseWriter, req *http.Request) {
		a.AuthRoutes.Verify(ctx, w, req)
	}

	// Wrapper para ValidateToken
	validateTokenHandler := func(w http.ResponseWriter, req *http.Request) {
		a.AuthRoutes.ValidateToken(ctx, w, req)
	}

	// Registra os manipuladores de rota
	http.HandleFunc("/registerlogin", registerLoginHandler)
	http.HandleFunc("/verify", verifyHandler)
	http.HandleFunc("/validatetoken", validateTokenHandler)

	// Inicia o servidor HTTP
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080" // Porta padrão se LISTEN_ADDR não for definido
	}
	log.Println("Listening on", addr)
	return http.ListenAndServe(addr, nil)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env:", err)
	}

	cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))

	client, err := config.ConnectDB(context.Background())
	if err != nil {
		log.Fatal("Erro ao conectar ao MongoDB:", err)
	}
	defer client.Disconnect(context.Background())

	// Função para criar a instância do componente api
	run := func(ctx context.Context, a *api) error { // Corrigida a assinatura da função
		return a.Start(ctx)
	}

	// Inicia o Service Weaver
	if err := weaver.Run(context.Background(), run); err != nil {
		log.Fatal(err)
	}
}
