package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/lib/pq"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

// -----------------------------------------------------------------------------
// CreateUser — INSERT + RETURNING, без sqlx, на чистом database/sql
// -----------------------------------------------------------------------------
func (r *UserRepo) CreateUser(ctx context.Context, u entity.User) (int64, error) {
	const q = `INSERT INTO users (email, pass_hash, created_at)
	           VALUES ($1, $2, NOW())
	           RETURNING id`

	var id int64
	if err := r.db.QueryRowContext(ctx, q, u.Email, u.PassHash).Scan(&id); err != nil {
		// ловим дубликат e-mail по SQLSTATE 23505
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, errs.ErrEmailTaken
		}
		return 0, err
	}
	return id, nil
}

// -----------------------------------------------------------------------------
// ByEmail — SELECT … WHERE email = $1
// -----------------------------------------------------------------------------
func (r *UserRepo) ByEmail(ctx context.Context, email string) (entity.User, error) {
	const q = `SELECT id, email, pass_hash, created_at
	           FROM   users
	           WHERE  email = $1`

	var u entity.User
	err := r.db.QueryRowContext(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.PassHash, &u.Created_at)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.User{}, errs.ErrUserNotFound
	}
	return u, err
}
