-- name: CreatePackage :exec
INSERT INTO work_packages(id,reference,title,description,initiative,kind,created_at,updated_at) VALUES(sqlc.arg(id),sqlc.arg(reference),sqlc.arg(title),sqlc.arg(description),sqlc.arg(initiative),sqlc.arg(kind),sqlc.arg(created_at),sqlc.arg(updated_at));

-- name: ListPackages :many
SELECT p.id,p.reference,p.title,p.description,p.initiative,p.kind,p.created_at,p.updated_at,COUNT(w.id)::int AS total,COUNT(w.id) FILTER (WHERE w.status='done')::int AS done,COUNT(w.id) FILTER (WHERE w.status='in_qa')::int AS in_qa FROM work_packages p LEFT JOIN work_items w ON w.package_id=p.id GROUP BY p.id ORDER BY p.created_at DESC;

-- name: GetPackage :one
SELECT p.id,p.reference,p.title,p.description,p.initiative,p.kind,p.created_at,p.updated_at,COUNT(w.id)::int AS total,COUNT(w.id) FILTER (WHERE w.status='done')::int AS done,COUNT(w.id) FILTER (WHERE w.status='in_qa')::int AS in_qa FROM work_packages p LEFT JOIN work_items w ON w.package_id=p.id WHERE p.id=sqlc.arg(id) GROUP BY p.id;
