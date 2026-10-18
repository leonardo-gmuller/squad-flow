package demand

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
)

type Metrics struct {
	Total                 int             `json:"total"`
	Active                int             `json:"active"`
	InQA                  int             `json:"in_qa"`
	Returned              int             `json:"returned"`
	Delivered             int             `json:"delivered"`
	FirstPassPercent      float64         `json:"first_pass_percent"`
	AverageLeadHours      float64         `json:"average_lead_hours"`
	ByKind                map[string]int  `json:"by_kind"`
	ByStatus              map[string]int  `json:"by_status"`
	ByPriority            map[string]int  `json:"by_priority"`
	Overdue               int             `json:"overdue"`
	Unassigned            int             `json:"unassigned"`
	TotalReturns          int             `json:"total_returns"`
	ReworkPercent         float64         `json:"rework_percent"`
	OnTimePercent         float64         `json:"on_time_percent"`
	DeliveriesWithDueDate int             `json:"deliveries_with_due_date"`
	Delivered30Days       int             `json:"delivered_30_days"`
	MedianLeadHours       float64         `json:"median_lead_hours"`
	P85LeadHours          float64         `json:"p85_lead_hours"`
	AverageActiveAgeHours float64         `json:"average_active_age_hours"`
	DailyDeliveries       []DailyDelivery `json:"daily_deliveries"`
}

type DailyDelivery struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func (s *UseCase) Metrics(ctx context.Context) (Metrics, error) {
	rows, err := s.queries.AllForMetrics(ctx)
	if err != nil {
		return Metrics{}, erring.Wrap(err, "demand.Metrics")
	}

	return CalculateMetrics(rows, s.now()), nil
}

// CalculateMetrics uses actual completed timestamps as the lead-time denominator.
// Due dates represent calendar days; a delivery on the due day is on time.
func CalculateMetrics(rows []entity.Demand, now time.Time) Metrics {
	metrics := Metrics{ByKind: map[string]int{}, ByStatus: map[string]int{}, ByPriority: map[string]int{}, DailyDeliveries: []DailyDelivery{}}
	passedFirst, leadHours := 0, float64(0)
	leadTimes := []float64{}
	reworked, onTime := 0, 0
	activeAge := float64(0)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	daily := map[string]int{}

	for _, item := range rows {
		metrics.Total++
		metrics.ByKind[string(item.Kind)]++
		metrics.ByStatus[string(item.Status)]++
		metrics.ByPriority[string(item.Priority)]++
		metrics.TotalReturns += item.ReturnCount

		if item.ReturnCount > 0 {
			reworked++
		}

		activeAge += recordOpenDemand(&metrics, item, now, today)

		if item.Status == entity.InQA {
			metrics.InQA++
		}

		if item.Status == entity.ChangesRequested {
			metrics.Returned++
		}

		if item.Status == entity.Done {
			metrics.Delivered++

			if item.ReturnCount == 0 {
				passedFirst++
			}

			observation := recordCompletion(&metrics, item, today, daily)
			if observation.hasTimestamp {
				leadHours += observation.leadHours
				leadTimes = append(leadTimes, observation.leadHours)
			}

			if observation.onTime {
				onTime++
			}
		}
	}

	if metrics.Delivered > 0 {
		metrics.FirstPassPercent = float64(passedFirst) * 100 / float64(metrics.Delivered)
	}

	if len(leadTimes) > 0 {
		metrics.AverageLeadHours = leadHours / float64(len(leadTimes))
		sort.Float64s(leadTimes)
		sampleCount := len(leadTimes)
		metrics.MedianLeadHours = leadTimes[sampleCount/2]

		if sampleCount%2 == 0 {
			metrics.MedianLeadHours = (leadTimes[sampleCount/2-1] + leadTimes[sampleCount/2]) / 2
		}

		metrics.P85LeadHours = leadTimes[int(math.Ceil(float64(sampleCount)*0.85))-1]
	}

	if metrics.Total > 0 {
		metrics.ReworkPercent = float64(reworked) * 100 / float64(metrics.Total)
	}

	if metrics.Active > 0 {
		metrics.AverageActiveAgeHours = activeAge / float64(metrics.Active)
	}

	if metrics.DeliveriesWithDueDate > 0 {
		metrics.OnTimePercent = float64(onTime) * 100 / float64(metrics.DeliveriesWithDueDate)
	}

	for i := 29; i >= 0; i-- {
		day := today.AddDate(0, 0, -i).Format("2006-01-02")
		metrics.DailyDeliveries = append(metrics.DailyDeliveries, DailyDelivery{day, daily[day]})
	}

	return metrics
}

type completionObservation struct {
	leadHours            float64
	hasTimestamp, onTime bool
}

func recordCompletion(metrics *Metrics, item entity.Demand, today time.Time, daily map[string]int) completionObservation {
	if item.CompletedAt == nil {
		return completionObservation{}
	}

	observation := completionObservation{leadHours: math.Max(0, item.CompletedAt.Sub(item.CreatedAt).Hours()), hasTimestamp: true}

	day := item.CompletedAt.UTC().Format("2006-01-02")
	if day >= today.AddDate(0, 0, -29).Format("2006-01-02") && day <= today.Format("2006-01-02") {
		metrics.Delivered30Days++
		daily[day]++
	}

	if item.DueDate != nil {
		metrics.DeliveriesWithDueDate++
		observation.onTime = day <= item.DueDate.Format("2006-01-02")
	}

	return observation
}

func recordOpenDemand(metrics *Metrics, item entity.Demand, now, today time.Time) float64 {
	if item.Status == entity.Done {
		return 0
	}

	metrics.Active++
	if item.Assignee == "" {
		metrics.Unassigned++
	}

	if item.DueDate != nil && item.DueDate.Format("2006-01-02") < today.Format("2006-01-02") {
		metrics.Overdue++
	}

	return math.Max(0, now.Sub(item.CreatedAt).Hours())
}
