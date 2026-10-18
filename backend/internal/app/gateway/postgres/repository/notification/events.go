package notificationrepo

import (
	"context"
	"fmt"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/erring"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/gateway/postgres/sqlc"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/ctxkey"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

// Events share the demand transaction, so failed or rolled-back writes create no notifications.
func insert(ctx context.Context, queries *sqlc.Queries, userID, kind, title, message, demandID string) error {
	if userID == "" {
		return nil
	}

	if actor, ok := ctxkey.GetUser(ctx); ok && actor.ID == userID {
		return nil
	}

	id, err := identifier.NewUUID()
	if err != nil {
		return erring.Wrap(err, "notificationrepo.insert.id")
	}
	// Admin accounts never receive notifications, even when assigned to a demand.
	err = queries.CreateNotification(ctx, sqlc.CreateNotificationParams{
		ID:       id,
		DemandID: demandID,
		Kind:     kind,
		Title:    title,
		Message:  message,
		UserID:   userID,
	})

	return erring.Wrap(err, "notificationrepo.insert")
}

func roles(item entity.Demand) map[string]string {
	result := map[string]string{}
	if item.AssigneeID != "" {
		result[item.AssigneeID] = "desenvolvimento"
	}

	if item.ReviewerID != "" {
		if result[item.ReviewerID] != "" {
			result[item.ReviewerID] += " e QA"
		} else {
			result[item.ReviewerID] = "QA"
		}
	}

	return result
}

func Assignments(ctx context.Context, queries *sqlc.Queries, before, after entity.Demand) error {
	oldRoles, newRoles := roles(before), roles(after)
	for userID, role := range newRoles {
		if oldRoles[userID] == role {
			continue
		}

		if err := insert(ctx, queries, userID, "assigned", "Nova atribuição", fmt.Sprintf("Você é responsável por %s em SF-%04d: %s.", role, after.Number, after.Title), after.ID); err != nil {
			return err
		}
	}

	for userID := range oldRoles {
		if newRoles[userID] != "" {
			continue
		}

		if err := insert(ctx, queries, userID, "unassigned", "Atribuição alterada", fmt.Sprintf("Você deixou de ser responsável por SF-%04d: %s.", after.Number, after.Title), after.ID); err != nil {
			return err
		}
	}

	return nil
}

func StatusChanged(ctx context.Context, queries *sqlc.Queries, before, after entity.Demand) error {
	if before.Status == after.Status {
		return nil
	}

	labels := map[entity.Status]string{entity.Backlog: "Backlog", entity.InProgress: "Em desenvolvimento", entity.InQA: "Em QA", entity.ChangesRequested: "Devolvida para correção", entity.Done: "Concluída"}

	title := "Etapa atualizada"
	if after.Status == entity.InQA {
		title = "Entrega enviada para QA"
	}

	if after.Status == entity.ChangesRequested {
		title = "Entrega devolvida para correção"
	}

	if after.Status == entity.Done {
		title = "Entrega concluída"
	}

	for userID := range roles(after) {
		if err := insert(ctx, queries, userID, "status_changed", title, fmt.Sprintf("SF-%04d: %s. Nova etapa: %s.", after.Number, after.Title, labels[after.Status]), after.ID); err != nil {
			return err
		}
	}

	return nil
}

func QAPassed(ctx context.Context, queries *sqlc.Queries, demandID, scenario string) error {
	row, err := queries.GetDemandNotificationRecipients(ctx, demandID)
	if err != nil {
		return erring.Wrap(err, "notificationrepo.QAPassed.demand")
	}

	item := entity.Demand{
		ID:         row.ID,
		Number:     row.Number,
		Title:      row.Title,
		AssigneeID: row.AssigneeID,
		ReviewerID: row.ReviewerID,
	}

	for userID := range roles(item) {
		if err := insert(ctx, queries, userID, "qa_passed", "Teste de QA aprovado", fmt.Sprintf("O cenário “%s” foi aprovado em SF-%04d: %s.", scenario, item.Number, item.Title), item.ID); err != nil {
			return err
		}
	}

	return nil
}
