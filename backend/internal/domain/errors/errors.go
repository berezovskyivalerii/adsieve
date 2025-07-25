package errs

import "errors"

var (
	ErrUserNotFound = errors.New("user with such credentials not found")
	ErrInvalidCreds = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrEmailTaken = errors.New("this email is already taken")
	ErrDuplicateClick = errors.New("click alredy registered")
	ErrClickNotFound = errors.New("click was not found")
	ErrConversionExists = errors.New("conversion already exists")
	ErrInvalidRange = errors.New("invalid date range")
	ErrNoAdAccess   = errors.New("no access to requested ad_id")
)