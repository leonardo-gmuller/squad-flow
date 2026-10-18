package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

// Decode accepts one JSON object, rejects unknown fields and limits request size.
func Decode(r *http.Request, value any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)

	var raw json.RawMessage

	if err := decoder.Decode(&raw); err != nil {
		return decodeError(err)
	}

	if len(raw) == 0 || raw[0] != '{' {
		return erring.ErrInvalidInput.WithMessage("Envie um objeto JSON válido.")
	}

	fields := json.NewDecoder(bytes.NewReader(raw))
	fields.DisallowUnknownFields()

	if err := fields.Decode(value); err != nil {
		return decodeError(err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return decodeError(err)
		}

		return erring.ErrInvalidInput.WithMessage("Envie apenas um objeto JSON.")
	}

	return nil
}

func decodeError(err error) error {
	var oversized *http.MaxBytesError
	if errors.As(err, &oversized) {
		return erring.ErrPayloadTooLarge
	}

	return erring.ErrInvalidInput.WithMessage("Envie um objeto JSON válido, com os campos esperados.")
}
