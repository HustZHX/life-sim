package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	TokenKindGate    = "gate"
	TokenKindAccess  = "access"
	TokenKindRefresh = "refresh"
)

var (
	ErrTokenInvalid = errors.New("token invalid")
	ErrTokenExpired = errors.New("token expired")
)

type Claims struct {
	Kind     string `json:"k"`
	UserID   string `json:"uid,omitempty"`
	Username string `json:"usr,omitempty"`
	Exp      int64  `json:"exp"`
}

type TokenIssuer struct {
	secret []byte
}

func NewTokenIssuer(secret string) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret)}
}

func (t *TokenIssuer) Sign(kind string, userID, username string, ttl time.Duration) (string, error) {
	claims := Claims{
		Kind:     kind,
		UserID:   userID,
		Username: username,
		Exp:      time.Now().Add(ttl).Unix(),
	}
	if kind == TokenKindGate {
		claims.UserID = ""
		claims.Username = ""
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	sig := t.sign(payload)
	return base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (t *TokenIssuer) Verify(token, expectedKind string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrTokenInvalid
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrTokenInvalid
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrTokenInvalid
	}
	if !hmac.Equal(sig, t.sign(payload)) {
		return nil, ErrTokenInvalid
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrTokenInvalid
	}
	if claims.Kind != expectedKind {
		return nil, ErrTokenInvalid
	}
	if time.Now().Unix() > claims.Exp {
		return nil, ErrTokenExpired
	}
	return &claims, nil
}

func (t *TokenIssuer) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, t.secret)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}
