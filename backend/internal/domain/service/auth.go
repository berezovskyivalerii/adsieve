package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/berezovskyivalerii/adsieve/internal/domain"
	"github.com/berezovskyivalerii/adsieve/internal/domain/entity"
	errs "github.com/berezovskyivalerii/adsieve/internal/domain/errors"
	"github.com/golang-jwt/jwt"
)

type PasswordHasher interface {
	Hash(pw string) (string, error)
	Compare(hash, plain string) error
}

type SessionsRepository interface {
	Create(ctx context.Context, token entity.RefreshSession) error
	Get(ctx context.Context, token string) (entity.RefreshSession, error)
}

type AuthService struct {
	repo        domain.AuthRepository
	sessionRepo SessionsRepository
	hasher      PasswordHasher
	ttl         time.Duration // Добавить ttl в конфиг
	jwtKey      []byte
}

func NewAuthService(r domain.AuthRepository, sessionrepo SessionsRepository, hasher PasswordHasher, key []byte) *AuthService {
	return &AuthService{repo: r, hasher: hasher, sessionRepo: sessionrepo, jwtKey: key, ttl: time.Hour * 24}
}

func (s *AuthService) SignUp(
	ctx context.Context,
	inp entity.SignInput,
) (accessToken, refreshToken string, err error) {
	if _, err := s.repo.ByEmail(ctx, inp.Email); err == nil {
		return "", "", errs.ErrEmailTaken
	}

	hash, err := s.hasher.Hash(inp.Password)
	if err != nil {
		return "", "", err
	}

	userID, err := s.repo.CreateUser(ctx, entity.User{
		Email:    inp.Email,
		PassHash: hash,
	})
	if err != nil {
		return "", "", err
	}

	return s.generateTokens(ctx, userID)
}

func (s *AuthService) SignIn(
	ctx context.Context,
	inp entity.SignInput,
) (string, string, error) {
	user, err := s.repo.ByEmail(ctx, inp.Email)
	if err != nil {
		return "", "", err
	}

	if err := s.hasher.Compare(user.PassHash, inp.Password); err != nil {
		return "", "", errs.ErrInvalidCreds
	}

	access, refresh, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *AuthService) generateTokens(ctx context.Context, userId int64) (string, string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   strconv.Itoa(int(userId)),
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(s.ttl).Unix(),
	})

	accessToken, err := t.SignedString(s.jwtKey)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := newRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err := s.sessionRepo.Create(ctx, entity.RefreshSession{
		UserID:    userId,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *AuthService) Refresh(ctx context.Context, oldRefresh string) (string, string, error) {
	session, err := s.sessionRepo.Get(ctx, oldRefresh)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errs.ErrInvalidRefreshToken
		}
		return "", "", err
	}

	if time.Now().After(session.ExpiresAt) {
		return "", "", errs.ErrRefreshTokenExpired
	}

	return s.generateTokens(ctx, session.UserID)
}
