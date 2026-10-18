-- name: ListWorkstreams :many
SELECT id,code,name,color,active,created_at,updated_at FROM workstreams ORDER BY created_at,code;

-- name: ListMembers :many
SELECT id,name,email FROM users ORDER BY lower(name),id;

-- name: CreateWorkstream :one
INSERT INTO workstreams(id,code,name,color,active) VALUES(sqlc.arg(id),sqlc.arg(code),sqlc.arg(name),sqlc.arg(color),sqlc.arg(active)) RETURNING id,code,name,color,active,created_at,updated_at;

-- name: UpdateWorkstream :one
UPDATE workstreams SET name=sqlc.arg(name),color=sqlc.arg(color),active=COALESCE(sqlc.narg(active)::boolean,active),updated_at=now() WHERE id=sqlc.arg(id) RETURNING id,code,name,color,active,created_at,updated_at;

-- name: GetWorkstreamActiveForShare :one
SELECT active FROM workstreams WHERE code=sqlc.arg(kind) FOR SHARE;

-- name: GetMemberNameForKeyShare :one
SELECT name FROM users WHERE id=sqlc.arg(id) FOR KEY SHARE;
