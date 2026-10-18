-- name: CreateReport :exec
INSERT INTO generated_reports(id,user_id,demand_id,prompt,payload,created_at) VALUES(sqlc.arg(id),sqlc.arg(user_id),NULLIF(sqlc.arg(demand_id)::text,'')::uuid,sqlc.arg(prompt),sqlc.arg(payload),sqlc.arg(created_at));

-- name: GetReport :one
SELECT payload,user_id FROM generated_reports WHERE id=sqlc.arg(id) AND user_id=sqlc.arg(user_id);

-- name: ListReports :many
SELECT ((payload - 'source_data') #- '{content,html}')::jsonb AS payload,user_id FROM generated_reports WHERE user_id=sqlc.arg(user_id) AND (sqlc.arg(demand_id)::text='' OR demand_id=NULLIF(sqlc.arg(demand_id)::text,'')::uuid) ORDER BY created_at DESC LIMIT 50;
