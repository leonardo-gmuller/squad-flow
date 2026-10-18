package demand

import (
	"testing"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

func TestMetricsDenominatorsDeadlinesAndPercentiles(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 13, 13, 0, 0, 0, time.UTC)
	done := now.Add(-24 * time.Hour)
	due := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	rows := []entity.Demand{
		{Status: entity.Done, CreatedAt: done.Add(-48 * time.Hour), CompletedAt: &done, DueDate: &due},
		{Status: entity.Done, CreatedAt: done.Add(-96 * time.Hour), CompletedAt: &done, ReturnCount: 2},
		{Status: entity.Done, CreatedAt: done.Add(-24 * time.Hour)},
		{Status: entity.InQA, CreatedAt: now.Add(-72 * time.Hour), DueDate: &due},
		{Status: entity.InProgress, CreatedAt: now.Add(-24 * time.Hour), DueDate: &now, Assignee: "Ana"},
	}

	metrics := CalculateMetrics(rows, now)
	if metrics.AverageLeadHours != 72 || metrics.MedianLeadHours != 72 || metrics.P85LeadHours != 96 {
		t.Fatalf("lead times: %+v", metrics)
	}

	if metrics.OnTimePercent != 100 || metrics.DeliveriesWithDueDate != 1 || metrics.Overdue != 1 {
		t.Fatalf("deadlines: %+v", metrics)
	}

	if metrics.TotalReturns != 2 || metrics.ReworkPercent != 20 || metrics.AverageActiveAgeHours != 48 || metrics.Unassigned != 1 || metrics.Delivered30Days != 2 {
		t.Fatalf("flow: %+v", metrics)
	}

	if len(metrics.DailyDeliveries) != 30 || metrics.DailyDeliveries[28].Count != 2 {
		t.Fatalf("daily: %+v", metrics.DailyDeliveries)
	}
}
