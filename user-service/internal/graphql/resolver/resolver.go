package resolver

import (
	"user-service/internal/infrastructure/persistence"

	"github.com/go-playground/validator/v10"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Validate   *validator.Validate
	Repository *persistence.UserRepository
}
