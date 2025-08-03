package main

import (
	"log"
	"net/http"
	"user-service/internal/database"
	"user-service/internal/graphql"
	"user-service/internal/graphql/generated"
	"user-service/internal/graphql/resolver"
	"user-service/pkg/config"
	logger "user-service/pkg/logger"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/vektah/gqlparser/v2/ast"
)

func main() {
	// Initialize logger
	logger.Init(config.GetConfig().Production, config.GetConfig().LogFilePath, zerolog.DebugLevel)
	database.Connect()
	defer database.Close()

	port := config.GetConfig().Port

	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{Validate: validator.New()}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(graphql.LoggerExtension{})
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
