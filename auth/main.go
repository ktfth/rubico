package main

import (
	"context"
	"log"

	"github.com/ServiceWeaver/weaver"
	"go.mongodb.org/mongo-driver/mongo"

	"auth/config"
	"auth/routes"
)

type api struct {
	weaver.Implements[weaver.Main]
	routes.AuthRoutes // routes.AuthRoutes
	mongodbClient     *mongo.Client
}

// Start is called by Service Weaver to start the API component.
func (a *api) Start(ctx context.Context) error {
	client, err := config.ConnectDB(ctx)
	if err != nil {
		return err
	}
	a.mongodbClient = client
	log.Println("Connected to MongoDB")
	return nil
}

func main() {
	run := func(ctx context.Context, a *api) error {
		// Aqui você pode adicionar qualquer lógica de inicialização adicional, se necessário.
		return nil
	}

	if err := weaver.Run(context.Background(), run); err != nil {
		log.Fatal(err)
	}
}
