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
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	PassHash  string    `json:"pass_hash" db:"pass_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SignInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=6"`
}

func (i *SignInput) Validate() error {
	return validate.Struct(i)
}
