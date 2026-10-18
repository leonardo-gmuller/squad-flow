package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/go-chi/chi/v5"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report"
	reportmock "github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/report/mocks"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
)

func TestReportQueriesUseAuthenticatedOwnerThroughUseCaseInterfaces(t *testing.T) {
	t.Parallel()

	user := entity.User{ID: "11111111-1111-4111-8111-111111111111", Role: "member"}

	resourceID := "22222222-2222-4222-8222-222222222222"
	for _, test := range []struct {
		name, path string
		expect     func(*reportmock.MockUseCase, *reportmock.MockJobUseCase)
	}{
		{name: "personal reports", path: "/api/v1/reports?demand_id=" + resourceID + "&user_id=forged", expect: func(reports *reportmock.MockUseCase, _ *reportmock.MockJobUseCase) {
			reports.EXPECT().List(gomock.Any(), user.ID, resourceID).Return([]report.Report{{ID: resourceID}}, nil)
		}},
		{name: "report detail", path: "/api/v1/reports/" + resourceID + "?user_id=forged", expect: func(reports *reportmock.MockUseCase, _ *reportmock.MockJobUseCase) {
			reports.EXPECT().Get(gomock.Any(), user.ID, resourceID).Return(report.Report{ID: resourceID}, nil)
		}},
		{name: "personal jobs", path: "/api/v1/reports/jobs?user_id=forged", expect: func(_ *reportmock.MockUseCase, jobs *reportmock.MockJobUseCase) {
			jobs.EXPECT().List(gomock.Any(), user.ID).Return([]report.Job{{ID: resourceID}}, nil)
		}},
		{name: "job detail", path: "/api/v1/reports/jobs/" + resourceID + "?user_id=forged", expect: func(_ *reportmock.MockUseCase, jobs *reportmock.MockJobUseCase) {
			jobs.EXPECT().Get(gomock.Any(), user.ID, resourceID).Return(report.Job{ID: resourceID}, nil)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			reports := reportmock.NewMockUseCase(ctrl)
			jobs := reportmock.NewMockJobUseCase(ctrl)
			test.expect(reports, jobs)

			router := chi.NewRouter()
			(&ReportHandler{Service: reports, Jobs: jobs}).Register(router)

			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			request = request.WithContext(ctxkey.PutUser(request.Context(), user))
			recorded := httptest.NewRecorder()
			router.ServeHTTP(recorded, request)

			if recorded.Code != http.StatusOK || !strings.Contains(recorded.Body.String(), resourceID) {
				t.Fatalf("incorrect personal history response: %d %s", recorded.Code, recorded.Body.String())
			}
		})
	}
}

func TestReportHandlerQueuesThroughUseCaseWithoutBuildingSynchronously(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	reports := reportmock.NewMockUseCase(ctrl)
	jobs := reportmock.NewMockJobUseCase(ctrl)
	user := entity.User{ID: "11111111-1111-4111-8111-111111111111", Role: "member"}
	key := "22222222-2222-4222-8222-222222222222"
	job := report.Job{ID: "33333333-3333-4333-8333-333333333333", Status: report.JobQueued}
	jobs.EXPECT().Submit(gomock.Any(), user, report.Request{Prompt: "Analise as entregas"}, key).Return(job, nil)

	router := chi.NewRouter()
	(&ReportHandler{Service: reports, Jobs: jobs, Enabled: true}).Register(router)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/reports/generate", strings.NewReader(`{"prompt":"Analise as entregas"}`))
	request.Header.Set("Idempotency-Key", key)
	request = request.WithContext(ctxkey.PutUser(request.Context(), user))
	recorded := httptest.NewRecorder()
	router.ServeHTTP(recorded, request)

	if recorded.Code != http.StatusAccepted || recorded.Header().Get("Location") != "/api/v1/reports/jobs/"+job.ID || !strings.Contains(recorded.Body.String(), `"status":"queued"`) {
		t.Fatalf("report was not queued: %d %s", recorded.Code, recorded.Body.String())
	}
}
