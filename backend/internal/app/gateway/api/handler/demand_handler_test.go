package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	demand "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	demandmock "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand/mocks"
)

func TestPostCreatesDemand(t *testing.T) {
	t.Parallel()

	mock := demandmock.NewMockUseCase(gomock.NewController(t))
	mock.EXPECT().Create(gomock.Any(), demand.CreateInput{Title: "Nova validação", Kind: entity.Improvement, Priority: "high"}).Return(entity.Demand{ID: "11111111-1111-4111-8111-111111111111", Number: 1, Title: "Nova validação", Kind: entity.Improvement, Priority: "high", Status: entity.Backlog}, nil)
	router := New(mock)
	body := `{"title":"Nova validação","kind":"improvement","priority":"high"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/demands", strings.NewReader(body))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("want HTTP 201, got %d: %s", res.Code, res.Body.String())
	}

	var output entity.Demand
	if err := json.Unmarshal(res.Body.Bytes(), &output); err != nil {
		t.Fatal(err)
	}

	if output.Number != 1 || output.Status != entity.Backlog {
		t.Fatalf("unexpected output: %+v", output)
	}
}

func TestRejectsUnknownInputFields(t *testing.T) {
	t.Parallel()

	mock := demandmock.NewMockUseCase(gomock.NewController(t))
	router := New(mock)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/demands", strings.NewReader(`{"title":"Nova validação","kind":"improvement","priority":"high","is_admin":true}`))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("unknown fields accepted: %d", res.Code)
	}
}

func TestConflictHTTPStatus(t *testing.T) {
	t.Parallel()

	mock := demandmock.NewMockUseCase(gomock.NewController(t))
	mock.EXPECT().Move(gomock.Any(), "id", demand.MoveInput{Status: entity.Done}).Return(entity.Demand{}, erring.ErrInvalidTransition)
	router := New(mock)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/demands/id/move", strings.NewReader(`{"status":"done"}`))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", res.Code)
	}
}

func TestSpreadsheetSafe(t *testing.T) {
	t.Parallel()

	for _, input := range []string{"=2+2", " +SUM(1,1)", "@cmd"} {
		if !strings.HasPrefix(spreadsheetSafe(input), "'") {
			t.Fatalf("dangerous formula %q not escaped", input)
		}
	}

	if spreadsheetSafe("normal value") != "normal value" {
		t.Fatal("safe text modified")
	}
}
