package graphql

import (
	"context"
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

	respHandler := next(ctx)

	return func(ctx context.Context) *graphql.Response {
		resp := respHandler(ctx)
		latency := time.Since(start)

		fields := []any{
			"operation", rc.OperationName,
			"operationType", string(rc.Operation.Operation),
			"latency", latency,
			"variables", rc.Variables,
			"errorCount", len(resp.Errors),
		}

		if len(resp.Errors) > 0 {
			fields = append(fields, "errors", resp.Errors)
		}

		logger.Info("GraphQL request processed", fields...)

		return resp
	}
}
