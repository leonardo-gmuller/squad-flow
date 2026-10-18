package security

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

func TestJWTRejectsExpiredWrongIssuerAndWrongAlgorithm(t *testing.T) {
	t.Parallel()

	tokens := Tokens{Secret: []byte("test-secret-that-is-longer-than-32-bytes")}
	session := entity.Session{ID: "11111111-1111-4111-8111-111111111111", UserID: "22222222-2222-4222-8222-222222222222"}

	raw, err := tokens.Sign(session, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	if id, testJWTRejectsExpiredWrongIssuerAndWrongAlgorithmErr := tokens.Verify(raw); testJWTRejectsExpiredWrongIssuerAndWrongAlgorithmErr != nil || id != session.ID {
		t.Fatalf("valid: %s %v", id, testJWTRejectsExpiredWrongIssuerAndWrongAlgorithmErr)
	}

	expired, _ := tokens.Sign(session, time.Now().Add(-time.Minute))
	if _, err = tokens.Verify(expired); err == nil {
		t.Fatal("accepted expired JWT")
	}

	for _, test := range []struct {
		method jwt.SigningMethod
		issuer string
	}{{jwt.SigningMethodHS256, "foreign"}, {jwt.SigningMethodHS384, "squadflow"}} {
		raw, _ := jwt.NewWithClaims(test.method, jwt.RegisteredClaims{ID: session.ID, Subject: session.UserID, Issuer: test.issuer, Audience: jwt.ClaimStrings{"squadflow-web"}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}).SignedString(tokens.Secret)
		if _, err = tokens.Verify(raw); err == nil {
			t.Fatal("accepted invalid issuer/algorithm")
		}
	}
}

func TestJWTRejectsMalformedResourceUUIDs(t *testing.T) {
	t.Parallel()

	tokens := Tokens{Secret: []byte("test-secret-that-is-longer-than-32-bytes")}
	for _, session := range []entity.Session{
		{ID: "session", UserID: "22222222-2222-4222-8222-222222222222"},
		{ID: "11111111-1111-4111-8111-111111111111", UserID: "123"},
	} {
		raw, err := tokens.Sign(session, time.Now().Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}

		if _, err = tokens.Verify(raw); err == nil {
			t.Fatal("accepted JWT with malformed resource UUID")
		}
	}
}
