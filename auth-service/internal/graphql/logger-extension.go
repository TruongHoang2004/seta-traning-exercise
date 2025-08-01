package graphql

import (
	"context"
	"log"
	"time"
	"user-service/pkg/logger"

	"github.com/99designs/gqlgen/graphql"
)

type LoggerExtension struct{}

func (l LoggerExtension) ExtensionName() string {
	return "LoggerExtension"
}

func (l LoggerExtension) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (l LoggerExtension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	start := time.Now()
	rc := graphql.GetOperationContext(ctx)

	// Lấy response handler
	respHandler := next(ctx)

	return func(ctx context.Context) *graphql.Response {
		resp := respHandler(ctx) // thực thi xử lý thật sự ở đây
		latency := time.Since(start)

		// Log với
		logger.Info("GraphQL request processed",
			"latency", latency.String(),
			"operation", rc.OperationName,
			"operationType", string(rc.Operation.Operation), // query / mutation
			"query", rc.RawQuery,
			"variables", rc.Variables,
			"errorCount", len(resp.Errors), // số lỗi (nếu có)
		)

		// In log bằng log.Printf (tuỳ mục đích)
		log.Printf("[GraphQL] %s: %s", rc.Operation.Operation, rc.RawQuery)
		log.Printf("[Variables]: %v", rc.Variables)

		return resp
	}
}
