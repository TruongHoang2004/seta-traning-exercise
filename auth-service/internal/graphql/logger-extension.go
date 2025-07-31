package graphql

import (
	"context"
	"log"
	"time"
	"user-service/pkg/logger"

	"github.com/99designs/gqlgen/graphql"
	"go.uber.org/zap"
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

		// Log với zap
		logger.Info("GraphQL request processed",
			zap.String("latency", latency.String()),
			zap.String("operation", rc.OperationName),
			zap.String("operationType", string(rc.Operation.Operation)), // query / mutation
			zap.String("query", rc.RawQuery),
			zap.Any("variables", rc.Variables),
			zap.Int("errorCount", len(resp.Errors)), // số lỗi (nếu có)
		)

		// In log bằng log.Printf (tuỳ mục đích)
		log.Printf("[GraphQL] %s: %s", rc.Operation.Operation, rc.RawQuery)
		log.Printf("[Variables]: %v", rc.Variables)

		return resp
	}
}
