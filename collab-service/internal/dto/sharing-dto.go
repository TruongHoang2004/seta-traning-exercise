package dto

import "collab-service/internal/models"

type ShareDTO struct {
	UserID     string            `json:"userId" binding:"required"`
	AccessRole models.AccessRole `json:"accessRole" binding:"required"`
}
