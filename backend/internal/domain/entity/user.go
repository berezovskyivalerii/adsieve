package entity

import (
	"github.com/go-playground/validator/v10"
	"time"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type User struct {
	UserID       int64     `json:"user_id" db:"user_id"`
	Email        string    `json:"email" db:"email"`
	PassHash     string    `json:"pass_hash" db:"pass_hash"`
	RegisteredAt time.Time `json:"registered_at" db:"registered_at"`
}

type SignInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6"`
}

func (i *SignInput) Validate() error {
	return validate.Struct(i)
}
