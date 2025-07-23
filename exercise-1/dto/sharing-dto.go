package dto

import "seta-training-exercise-1/models"

type ShareDTO struct {
	UserID     string            `json:"userId" binding:"required"`
	AccessRole models.AccessRole `json:"accessRole" binding:"required"`
}
