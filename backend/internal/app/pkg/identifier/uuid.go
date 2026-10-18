// Package identifier centralizes UUID generation and validation for application resources.
package identifier

import (
	"github.com/google/uuid"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

// NewUUID generates a random UUID v4 in the canonical JSON/URL representation.
func NewUUID() (string, error) {
	value, err := uuid.NewRandom()
	if err != nil {
		return "", erring.Wrap(err, "identifier.NewUUID")
	}

	return value.String(), nil
}

// IsUUID accepts the existing 36-character API format, including uppercase hex.
// Google UUID also accepts URNs, braces and unhyphenated values; those are not API identifiers.
func IsUUID(value string) bool { return len(value) == 36 && uuid.Validate(value) == nil }
