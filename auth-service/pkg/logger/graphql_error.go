package logger

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// NewUserError tạo lỗi GraphQL đại diện cho lỗi do người dùng gây ra (ví dụ: input sai).
func NewUserError(message string, code string) *gqlerror.Error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
}

// NewSystemError tạo lỗi GraphQL đại diện cho lỗi hệ thống (DB lỗi, panic, internal).
// Đồng thời log lỗi với cấp độ Error.
func NewSystemError(message string, fields ...any) *gqlerror.Error {
	Error(message, fields...) // ghi log luôn
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code": "INTERNAL_SERVER_ERROR",
		},
	}
}

// FromError chuyển một lỗi Go thông thường thành lỗi GraphQL.
// Dựa vào cờ `isSystem` để xác định loại lỗi.
func FromError(err error, isSystem bool) *gqlerror.Error {
	if isSystem {
		Error("System error: "+err.Error(), "error", err)
		return &gqlerror.Error{
			Message: "Internal server error",
			Extensions: map[string]interface{}{
				"code": "INTERNAL_SERVER_ERROR",
			},
		}
	}

	return &gqlerror.Error{
		Message: err.Error(),
		Extensions: map[string]interface{}{
			"code": "BAD_USER_INPUT",
		},
	}
}

// WrapResolverError dùng để đính kèm lỗi vào context GraphQL hiện tại,
// giữ nguyên lỗi gốc nhưng thêm `extensions.code` để frontend xử lý dễ hơn.
func WrapResolverError(ctx context.Context, err error, code string) *gqlerror.Error {
	gqlErr := graphql.DefaultErrorPresenter(ctx, err)
	if gqlErr.Extensions == nil {
		gqlErr.Extensions = make(map[string]interface{})
	}
	gqlErr.Extensions["code"] = code
	return gqlErr
}
