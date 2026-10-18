package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

type Passwords struct{}

func (Passwords) Hash(raw string) (string, error) {
	b, e := bcrypt.GenerateFromPassword([]byte(raw), 12)

	return string(b), e
}

func (Passwords) Match(hash, raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}

type Tokens struct{ Secret []byte }

func (t Tokens) NewID() (string, error) {
	result, queryErr := identifier.NewUUID()

	return result, erring.Wrap(queryErr, "security.NewID")
}

func (t Tokens) Refresh() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", erring.Wrap(err, "security.Refresh")
	}

	raw := base64.RawURLEncoding.EncodeToString(b)

	return raw, t.HashRefresh(raw), nil
}

func (Tokens) HashRefresh(raw string) string {
	hash := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(hash[:])
}

func (t Tokens) Sign(s entity.Session, expires time.Time) (string, error) {
	result, queryErr := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Issuer: "squadflow", Audience: jwt.ClaimStrings{"squadflow-web"}, Subject: s.UserID, ID: s.ID, IssuedAt: jwt.NewNumericDate(time.Now()), ExpiresAt: jwt.NewNumericDate(expires)}).SignedString(t.Secret)

	return result, erring.Wrap(queryErr, "security.Sign")
}

func (t Tokens) Verify(raw string) (string, error) {
	claims := new(jwt.RegisteredClaims)
	token, err := jwt.ParseWithClaims(raw, claims, func(_ *jwt.Token) (any, error) { return t.Secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("squadflow"), jwt.WithAudience("squadflow-web"), jwt.WithExpirationRequired())

	if err != nil || !token.Valid || !identifier.IsUUID(claims.ID) || !identifier.IsUUID(claims.Subject) {
		return "", errors.New("invalid access token")
	}

	return claims.ID, nil
}
