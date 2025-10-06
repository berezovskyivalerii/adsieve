package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) CreateUser(ctx context.Context, u entity.User) (int64, error) {
	const q = `INSERT INTO users (email, password_hash, registered_at)
	           VALUES ($1, $2, NOW())
	           RETURNING user_id`

	var user_id int64
	if err := r.db.QueryRowContext(ctx, q, u.Email, u.PassHash).Scan(&user_id); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, errs.ErrEmailTaken
		}
		return 0, err
	}

	return user_id, nil
}

func (r *UserRepo) ByEmail(ctx context.Context, email string) (entity.User, error) {
	const q = `SELECT user_id, email, password_hash, registered_at
	           FROM   users
	           WHERE  email = $1`

	var u entity.User
	err := r.db.QueryRowContext(ctx, q, email).
		Scan(&u.UserID, &u.Email, &u.PassHash, &u.RegisteredAt)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, errs.ErrUserNotFound
	}
	return u, err
}
