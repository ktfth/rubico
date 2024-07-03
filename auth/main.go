package main

import (
	"auth/routes"
	"context"
	_ "embed"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ServiceWeaver/weaver"
	"go.mongodb.org/mongo-driver/mongo"

	"auth/config"
	_ "auth/docs"
)

// swagger embed files

// api is the main application component.
type api struct {
	weaver.Implements[weaver.Main]
	routes.AuthRoutes
	mongodbClient *mongo.Client
}

////go:embed frontend/index.html
//var indexHtml string

// Start is called by Service Weaver to start the API component.
func (a *api) Start(ctx context.Context) error {
	client, err := config.ConnectDB(ctx)
	if err != nil {
		return err
	}
	a.mongodbClient = client
	log.Println("Connected to MongoDB")

	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Executar a cada hora (ajuste conforme necessário)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				// O contexto foi cancelado, encerrar a goroutine
				return
			case <-ticker.C:
				// Verificar e remover tokens expirados do banco de dados
				if err := expireTokens(ctx, a.mongodbClient); err != nil {
					log.Printf("Erro ao expirar tokens: %v", err)
				}
			}
		}
	}()

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
	http.Handle("/docs", http.FileServer(http.Dir("./docs")))
	http.Handle(
		"/swagger/",
		httpSwagger.WrapHandler,
	)
	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	// Inicia o servidor HTTP
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080" // Porta padrão se LISTEN_ADDR não for definido
	}
	log.Println("Listening on", addr)
	return http.ListenAndServe(addr, nil)
}

func expireTokens(ctx context.Context, client *mongo.Client) error {
	// 1. Obter a coleção de tokens
	collection := client.Database("rubico").Collection("tokens")

	// 2. Definir o filtro para encontrar tokens expirados
	filter := bson.M{"expiresAt": bson.M{"$lt": time.Now()}}

	// 3. Remover os tokens expirados
	result, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	log.Printf("Tokens expirados removidos: %d", result.DeletedCount)
	return nil
}

// @title           Rubico API
// @version         0.0
// @description     Magic Authentication
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    https://kaeyosthaeron.com
// @contact.email  kaeyosthaeron@gmailc.om

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
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
