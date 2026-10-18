//go:build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/entity"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/domain/usecase/demand"
	"github.com/leonardo-gmuller/squad-flow/backend/internal/app/pkg/identifier"
)

func assertCatalogEndpoints(t *testing.T, admin, member *http.Client, serverURL, userID string) {
	t.Helper()

	call := func(client *http.Client, method, path string, input any, status int) []byte {
		return callAPI(t, client, serverURL, method, path, input, status)
	}
	call(testClient(t), "GET", "/api/v1/workstreams", nil, 401)
	call(testClient(t), "GET", "/api/v1/members", nil, 401)

	members := call(member, "GET", "/api/v1/members", nil, 200)
	if !bytes.Contains(members, []byte(userID)) || bytes.Contains(members, []byte("password")) {
		t.Fatal("unsafe or incomplete member directory")
	}

	id, _ := identifier.NewUUID()
	input := map[string]any{"name": "Infra " + id, "color": "#2c8c99", "active": true}
	call(member, "POST", "/api/v1/workstreams", input, 403)
	raw := call(admin, "POST", "/api/v1/workstreams", input, 201)

	var stream entity.Workstream

	if err := json.Unmarshal(raw, &stream); err != nil {
		t.Fatal(err)
	}

	if stream.Code == "" || !stream.Active {
		t.Fatal("invalid workstream")
	}

	input["name"] = "  " + strings.ToUpper(stream.Name) + "  "
	call(admin, "POST", "/api/v1/workstreams", input, 409)
	input["name"] = stream.Name
	input["color"] = "red"
	call(admin, "POST", "/api/v1/workstreams", input, 400)
	input["color"] = "#2c8c99"
	call(member, "PUT", "/api/v1/workstreams/"+stream.ID, input, 403)
	create := map[string]any{"title": "Demanda com esteira cadastrada", "kind": stream.Code, "priority": "medium", "assignee_id": userID, "reviewer_id": userID}
	raw = call(member, "POST", "/api/v1/demands", create, 201)

	var item entity.Demand

	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatal(err)
	}

	if item.AssigneeID != userID || item.ReviewerID != userID || item.Assignee == "" {
		t.Fatal("assignment did not resolve registered users")
	}

	create["assignee_id"] = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	call(member, "POST", "/api/v1/demands", create, 400)
	create["assignee_id"] = "nome livre"
	call(member, "POST", "/api/v1/demands", create, 400)
	create["assignee_id"] = userID
	create["assignee"] = "nome livre"
	call(member, "POST", "/api/v1/demands", create, 400)
	delete(create, "assignee")
	create["kind"] = "unknown"
	call(member, "POST", "/api/v1/demands", create, 400)
	create["kind"] = stream.Code
	pkg := map[string]any{"title": "OS com esteira cadastrada", "kind": stream.Code, "create_first_item": true, "assignee_id": userID, "reviewer_id": userID}
	call(member, "POST", "/api/v1/packages", pkg, 201)

	var board demand.Board
	if err := json.Unmarshal(call(member, "GET", "/api/v1/board?kind="+stream.Code+"&owner="+userID, nil, 200), &board); err != nil {
		t.Fatal(err)
	}

	if len(board.Columns) != 5 || board.Columns[0].Total != 2 {
		t.Fatal("custom workstream or user filter failed")
	}

	input["name"] = "Renomeada " + id
	raw = call(admin, "PUT", "/api/v1/workstreams/"+stream.ID, input, 200)

	var renamed entity.Workstream

	if err := json.Unmarshal(raw, &renamed); err != nil || renamed.Code != stream.Code {
		t.Fatal("renaming broke category identity")
	}

	input["active"] = false
	call(admin, "PUT", "/api/v1/workstreams/"+stream.ID, input, 200)
	call(member, "POST", "/api/v1/demands", create, 409)
	call(member, "POST", "/api/v1/packages", pkg, 409)

	edit := map[string]any{"title": "Título atualizado com histórico preservado", "priority": "high", "assignee_id": userID, "reviewer_id": userID}
	call(member, "PUT", "/api/v1/demands/"+item.ID, edit, 200)
	edit["reviewer_id"] = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	call(member, "PUT", "/api/v1/demands/"+item.ID, edit, 400)

	for _, status := range []string{"in_progress", "in_qa"} {
		call(member, "POST", "/api/v1/demands/"+item.ID+"/move", map[string]string{"status": status, "actor": "forged"}, 200)
	}

	call(member, "POST", "/api/v1/demands/"+item.ID+"/tests", map[string]string{"title": "Teste de usuário registrado", "result": "passed", "author": "forged"}, 201)

	var detail demand.Detail
	if err := json.Unmarshal(call(member, "GET", "/api/v1/demands/"+item.ID, nil, 200), &detail); err != nil {
		t.Fatal(err)
	}

	if detail.Demand.Kind != entity.Kind(stream.Code) || detail.Demand.ReviewerID != userID || len(detail.Transitions) != 2 || len(detail.Tests) != 1 {
		t.Fatal("history or assignment was lost")
	}

	for _, transition := range detail.Transitions {
		if transition.ActorID != userID || transition.Actor == "forged" {
			t.Fatal("unregistered actor recorded")
		}
	}

	if detail.Tests[0].AuthorID != userID || detail.Tests[0].Author == "forged" {
		t.Fatal("unregistered test author recorded")
	}

	input["active"] = true
	call(admin, "PUT", "/api/v1/workstreams/"+stream.ID, input, 200)
	call(member, "POST", "/api/v1/demands", create, 201)
}
