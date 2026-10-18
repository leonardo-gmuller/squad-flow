package workpackage

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
)

func TestCreatePackageAndFirstItemUseTheSameTransaction(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name         string
		packageError error
		itemError    error
	}{
		{name: "success"},
		{name: "package insert fails", packageError: errors.New("package insertion failed")},
		{name: "first item insert fails", itemError: errors.New("item insertion failed")},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			uow := NewMockUnitOfWork(ctrl)
			queries := NewMockRepository(ctrl)
			repos := NewMockRepositories(ctrl)
			packages := NewMockRepository(ctrl)
			demands := NewMockDemandRepository(ctrl)
			packageID := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
			itemID := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
			assigneeID := "55555555-5555-4555-8555-555555555501"
			transactionCtx := context.WithValue(t.Context(), transactionKey{}, "transaction")

			uow.EXPECT().DoPackage(t.Context(), gomock.Any()).DoAndReturn(func(_ context.Context, callback func(context.Context, Repositories) error) error {
				return callback(transactionCtx, repos)
			})

			packageRepository := repos.EXPECT().Packages().Return(packages)

			packageInsert := packages.EXPECT().Insert(transactionCtx, gomock.Cond(func(item entity.WorkPackage) bool {
				return item.ID == packageID && item.Title == "Implementar regra legal" && item.Reference == "OS-001" && item.Kind == entity.Improvement
			})).Return(test.packageError).After(packageRepository)
			if test.packageError == nil {
				demandRepository := repos.EXPECT().Demands().Return(demands).After(packageInsert)
				demands.EXPECT().Insert(transactionCtx, gomock.Cond(func(item *entity.Demand) bool {
					return item.ID == itemID && item.PackageID == packageID && item.AssigneeID == assigneeID && item.Status == entity.Backlog && item.Title == "Implementar regra legal"
				})).Return(test.itemError).After(demandRepository)
			}

			next := 0
			service := New(uow, queries, func() (string, error) {
				next++
				if next == 1 {
					return packageID, nil
				}

				return itemID, nil
			})
			workPackage, err := service.Create(t.Context(), Input{Reference: " OS-001 ", Title: " Implementar regra legal ", Kind: entity.Improvement, CreateFirstItem: true, AssigneeID: assigneeID})

			expectedError := test.packageError
			if expectedError == nil {
				expectedError = test.itemError
			}

			if expectedError != nil {
				if !errors.Is(err, expectedError) || workPackage.ID != "" {
					t.Fatalf("failed transaction returned a created package: package=%+v error=%v", workPackage, err)
				}

				return
			}

			if err != nil || workPackage.ID != packageID || workPackage.Total != 1 {
				t.Fatalf("unexpected package: %+v error=%v", workPackage, err)
			}
		})
	}
}

type transactionKey struct{}
