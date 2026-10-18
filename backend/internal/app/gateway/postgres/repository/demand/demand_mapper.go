package demandrepo

import (
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
)

func toDemand(row sqlc.GetDemandRow) entity.Demand {
	return entity.Demand{
		ID:                 row.ID,
		Number:             row.Number,
		Title:              row.Title,
		Description:        row.Description,
		Kind:               entity.Kind(row.Kind),
		Priority:           entity.Priority(row.Priority),
		Status:             entity.Status(row.Status),
		Assignee:           row.Assignee,
		Reviewer:           row.Reviewer,
		Reference:          row.Reference,
		EventCodes:         row.EventCodes,
		QACycle:            row.QACycle,
		ReturnCount:        row.ReturnCount,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
		CompletedAt:        row.CompletedAt,
		PackageID:          row.PackageID,
		Product:            row.Product,
		AcceptanceCriteria: row.AcceptanceCriteria,
		DueDate:            row.DueDate,
		ReleaseID:          row.ReleaseID,
		ReleaseName:        row.ReleaseName,
		AssigneeID:         row.AssigneeID,
		ReviewerID:         row.ReviewerID,
		KindName:           row.KindName,
	}
}

func toTestCase(row sqlc.ListDemandTestsRow) entity.TestCase {
	return entity.TestCase{
		ID:        row.ID,
		DemandID:  row.DemandID,
		Cycle:     row.Cycle,
		Title:     row.Title,
		Steps:     row.Steps,
		Expected:  row.Expected,
		Actual:    row.Actual,
		Result:    row.Result,
		Author:    row.Author,
		CreatedAt: row.CreatedAt,
		AuthorID:  row.AuthorID,
	}
}

func toTransition(row sqlc.ListDemandTransitionsRow) entity.Transition {
	return entity.Transition{
		ID:        row.ID,
		DemandID:  row.DemandID,
		From:      entity.Status(row.FromStatus),
		To:        entity.Status(row.ToStatus),
		Actor:     row.Actor,
		Reason:    row.Reason,
		CreatedAt: row.CreatedAt,
		ActorID:   row.ActorID,
	}
}

func createDemandParams(item entity.Demand) sqlc.CreateDemandParams {
	return sqlc.CreateDemandParams{
		ID:                 item.ID,
		Title:              item.Title,
		Description:        item.Description,
		Kind:               string(item.Kind),
		Priority:           string(item.Priority),
		Status:             string(item.Status),
		Assignee:           item.Assignee,
		Reviewer:           item.Reviewer,
		Reference:          item.Reference,
		EventCodes:         item.EventCodes,
		QACycle:            item.QACycle,
		ReturnCount:        item.ReturnCount,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
		PackageID:          item.PackageID,
		Product:            item.Product,
		AcceptanceCriteria: item.AcceptanceCriteria,
		DueDate:            item.DueDate,
		ReleaseID:          item.ReleaseID,
		AssigneeID:         item.AssigneeID,
		ReviewerID:         item.ReviewerID,
	}
}

func updateDemandDetailsParams(item entity.Demand) sqlc.UpdateDemandDetailsParams {
	return sqlc.UpdateDemandDetailsParams{
		Title:              item.Title,
		Description:        item.Description,
		Priority:           string(item.Priority),
		Assignee:           item.Assignee,
		Reviewer:           item.Reviewer,
		Product:            item.Product,
		Reference:          item.Reference,
		AcceptanceCriteria: item.AcceptanceCriteria,
		DueDate:            item.DueDate,
		EventCodes:         item.EventCodes,
		UpdatedAt:          item.UpdatedAt,
		ReleaseID:          item.ReleaseID,
		AssigneeID:         item.AssigneeID,
		ReviewerID:         item.ReviewerID,
		ID:                 item.ID,
	}
}

func updateDemandStatusParams(item entity.Demand) sqlc.UpdateDemandStatusParams {
	return sqlc.UpdateDemandStatusParams{
		Status:      string(item.Status),
		QACycle:     item.QACycle,
		ReturnCount: item.ReturnCount,
		UpdatedAt:   item.UpdatedAt,
		CompletedAt: item.CompletedAt,
		ID:          item.ID,
	}
}
