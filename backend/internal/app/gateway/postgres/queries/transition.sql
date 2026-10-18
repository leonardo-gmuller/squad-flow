-- name: CreateTransition :exec
INSERT INTO work_item_transitions(id,demand_id,from_status,to_status,actor,reason,created_at,actor_id) VALUES(sqlc.arg(id),sqlc.arg(demand_id),sqlc.arg(from_status),sqlc.arg(to_status),sqlc.arg(actor),sqlc.arg(reason),sqlc.arg(created_at),NULLIF(sqlc.arg(actor_id)::text,'')::uuid);
