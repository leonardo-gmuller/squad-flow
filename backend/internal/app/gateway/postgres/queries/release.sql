-- name: ListReleases :many
SELECT r.id,r.name,r.description,r.release_date,r.archived_at,r.created_at,r.updated_at,
 count(w.id)::int AS total,count(w.id) FILTER (WHERE w.status='done')::int AS done,count(w.id) FILTER (WHERE w.status='in_qa')::int AS in_qa
 FROM releases r LEFT JOIN work_items w ON w.release_id=r.id GROUP BY r.id ORDER BY (r.archived_at IS NOT NULL),r.created_at DESC,r.id;

-- name: CreateRelease :exec
INSERT INTO releases(id,name,description,release_date,created_at,updated_at) VALUES(sqlc.arg(id),sqlc.arg(name),sqlc.arg(description),sqlc.arg(release_date),sqlc.arg(created_at),sqlc.arg(updated_at));

-- name: UpdateRelease :execrows
UPDATE releases SET name=sqlc.arg(name),description=sqlc.arg(description),release_date=sqlc.arg(release_date),updated_at=sqlc.arg(updated_at) WHERE id=sqlc.arg(id);

-- name: SetReleaseArchived :execrows
UPDATE releases SET archived_at=CASE WHEN sqlc.arg(archived)::boolean THEN COALESCE(archived_at,sqlc.arg(updated_at)) ELSE NULL END,updated_at=sqlc.arg(updated_at) WHERE id=sqlc.arg(id);
