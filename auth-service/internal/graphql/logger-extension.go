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

		hasSystemError := false
		hasUserError := false

		for _, err := range resp.Errors {
			if code, ok := err.Extensions["code"].(string); ok {
				switch code {
				case "INTERNAL_SERVER_ERROR", "DATABASE_ERROR", "PANIC":
					hasSystemError = true
				case "BAD_USER_INPUT", "UNAUTHORIZED", "FORBIDDEN", "NOT_FOUND":
					hasUserError = true
				}
			} else {
				// Nếu không có extensions.code => xem như lỗi hệ thống
				hasSystemError = true
			}
		}

		fields := []any{
			"operation", rc.OperationName,
			"operationType", string(rc.Operation.Operation),
			"latency", latency.String(),
			"variables", rc.Variables,
			"errorCount", len(resp.Errors),
		}

		if len(resp.Errors) > 0 {
			fields = append(fields, "errors", resp.Errors)
		}

		switch {
		case hasSystemError:
			fields = append(fields, "status", "system_error")
			logger.Error("GraphQL system error", fields...)
		case hasUserError:
			fields = append(fields, "status", "user_error")
			logger.Warn("GraphQL user error", fields...)
		default:
			fields = append(fields, "status", "success")
			logger.Info("GraphQL request success", fields...)
		}

		return resp
	}
}
