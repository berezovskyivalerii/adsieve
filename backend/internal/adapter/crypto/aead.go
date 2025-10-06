package crypto

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// AEADEncryptor реализует симметричное шифрование AES-256-GCM.
// Требуется ключ длиной 32 байта (256 бит). Ключ можно передать:
// - как base64 (StdEncoding или RawStdEncoding),
// - как hex,
// - как "сырую" строку длиной ровно 32 байта.
type AEADEncryptor struct {
	aead cipher.AEAD
}

// NewAEADEncryptor создаёт шифратор из строкового ключа.
// Возвращает ошибку, если не удалось распарсить или длина ключа != 32 байта.
func NewAEADEncryptor(keyMaterial string) (*AEADEncryptor, error) {
	key, err := parseKey32(keyMaterial)
	if err != nil {
		return nil, fmt.Errorf("aead: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aead: new cipher: %w", err)
	}
	a, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aead: new gcm: %w", err)
	}
	return &AEADEncryptor{aead: a}, nil
}

// EncryptString шифрует plain и возвращает base64url строку вида v1:<nonce+ciphertext>.
// Nonce генерируется случайно (aead.NonceSize()).
func (e *AEADEncryptor) EncryptString(_ context.Context, plain string) (string, error) {
	if e == nil || e.aead == nil {
		return "", errors.New("aead: not initialized")
	}
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("aead: nonce: %w", err)
	}
	ct := e.aead.Seal(nil, nonce, []byte(plain), nil)

	out := make([]byte, 0, len(nonce)+len(ct))
	out = append(out, nonce...)
	out = append(out, ct...)

	// base64 URL-safe без '='
	return "v1:" + base64.RawURLEncoding.EncodeToString(out), nil
}

// DecryptString расшифровывает строку, сгенерированную EncryptString.
func (e *AEADEncryptor) DecryptString(_ context.Context, enc string) (string, error) {
	if e == nil || e.aead == nil {
		return "", errors.New("aead: not initialized")
	}
	const prefix = "v1:"
	if len(enc) < len(prefix) || enc[:len(prefix)] != prefix {
		return "", errors.New("aead: bad prefix")
	}
	raw, err := base64.RawURLEncoding.DecodeString(enc[len(prefix):])
	if err != nil {
		return "", fmt.Errorf("aead: b64: %w", err)
	}
	ns := e.aead.NonceSize()
	if len(raw) < ns {
		return "", errors.New("aead: short payload")
	}
	nonce, ct := raw[:ns], raw[ns:]
	pt, err := e.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("aead: open: %w", err)
	}
	return string(pt), nil
}

// ===== helpers =====

func parseKey32(s string) ([]byte, error) {
	if s == "" {
		return nil, errors.New("empty key")
	}
	// 1) base64 / base64url (с '=' и без)
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	if b, err := base64.RawStdEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	if b, err := base64.URLEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	// 2) hex
	if b, err := hex.DecodeString(s); err == nil && len(b) == 32 {
		return b, nil
	}
	// 3) raw (ровно 32 байта)
	if len(s) == 32 {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("invalid key length (need 32 bytes after decode)")
}
