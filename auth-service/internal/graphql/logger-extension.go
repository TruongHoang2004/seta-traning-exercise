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
				hasSystemError = true
			}
		}

		// List of sensitive keys to redact
		sensitiveKeys := map[string]struct{}{
			"password":        {},
			"newPassword":     {},
			"currentPassword": {},
			"confirmPassword": {},
			"token":           {},
			"accessToken":     {},
			"refreshToken":    {},
		}

		filteredVars := sanitizeVariables(rc.Variables, sensitiveKeys)

		fields := []any{
			"operation", rc.OperationName,
			"operationType", string(rc.Operation.Operation),
			"latency", latency.String(),
			"variables", filteredVars,
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

func sanitizeVariables(vars map[string]interface{}, sensitiveKeys map[string]struct{}) map[string]interface{} {
	safe := make(map[string]interface{}, len(vars))

	for k, v := range vars {
		if _, sensitive := sensitiveKeys[k]; sensitive {
			safe[k] = "[REDACTED]"
			continue
		}

		// Nếu là map lồng, gọi đệ quy
		if nestedMap, ok := v.(map[string]interface{}); ok {
			safe[k] = sanitizeVariables(nestedMap, sensitiveKeys)
		} else {
			safe[k] = v
		}
	}

	return safe
}
