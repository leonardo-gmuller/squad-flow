-- name: CreateDemand :one
INSERT INTO work_items(id,title,description,kind,priority,status,assignee,reviewer,reference,event_codes,qa_cycle,return_count,created_at,updated_at,package_id,product,acceptance_criteria,due_date,release_id,assignee_id,reviewer_id)
        VALUES(sqlc.arg(id),sqlc.arg(title),sqlc.arg(description),sqlc.arg(kind)::text,sqlc.arg(priority),sqlc.arg(status)::text,sqlc.arg(assignee),sqlc.arg(reviewer),sqlc.arg(reference),sqlc.arg(event_codes),sqlc.arg(qa_cycle),sqlc.arg(return_count),sqlc.arg(created_at),sqlc.arg(updated_at),NULLIF(sqlc.arg(package_id)::text,'')::uuid,sqlc.arg(product),sqlc.arg(acceptance_criteria),sqlc.arg(due_date),NULLIF(sqlc.arg(release_id)::text,'')::uuid,NULLIF(sqlc.arg(assignee_id)::text,'')::uuid,NULLIF(sqlc.arg(reviewer_id)::text,'')::uuid) RETURNING number;

-- name: GetDemandForUpdate :one
SELECT id, number, title, description, kind, priority, status, COALESCE((SELECT name FROM users WHERE users.id=work_items.assignee_id),assignee)::text AS assignee, COALESCE((SELECT name FROM users WHERE users.id=work_items.reviewer_id),reviewer)::text AS reviewer, reference, event_codes, qa_cycle, return_count, created_at, updated_at, completed_at, COALESCE(package_id::text,'')::text AS package_id, product, acceptance_criteria, due_date, COALESCE(release_id::text,'')::text AS release_id, COALESCE((SELECT name FROM releases WHERE releases.id=work_items.release_id),'')::text AS release_name, COALESCE(assignee_id::text,'')::text AS assignee_id, COALESCE(reviewer_id::text,'')::text AS reviewer_id, COALESCE((SELECT name FROM workstreams WHERE workstreams.code=work_items.kind),'')::text AS kind_name FROM work_items WHERE work_items.id=sqlc.arg(id) FOR UPDATE;

-- name: UpdateDemandStatus :execrows
UPDATE work_items SET status=sqlc.arg(status)::text,qa_cycle=sqlc.arg(qa_cycle),return_count=sqlc.arg(return_count),updated_at=sqlc.arg(updated_at),completed_at=sqlc.arg(completed_at) WHERE id=sqlc.arg(id);

-- name: UpdateDemandDetails :execrows
UPDATE work_items SET title=sqlc.arg(title),description=sqlc.arg(description),priority=sqlc.arg(priority),assignee=sqlc.arg(assignee),reviewer=sqlc.arg(reviewer),product=sqlc.arg(product),reference=sqlc.arg(reference),acceptance_criteria=sqlc.arg(acceptance_criteria),due_date=sqlc.arg(due_date),event_codes=sqlc.arg(event_codes),updated_at=sqlc.arg(updated_at),release_id=NULLIF(sqlc.arg(release_id)::text,'')::uuid,assignee_id=NULLIF(sqlc.arg(assignee_id)::text,'')::uuid,reviewer_id=NULLIF(sqlc.arg(reviewer_id)::text,'')::uuid WHERE id=sqlc.arg(id);

-- name: ListDemands :many
SELECT id, number, title, description, kind, priority, status, COALESCE((SELECT name FROM users WHERE users.id=work_items.assignee_id),assignee)::text AS assignee, COALESCE((SELECT name FROM users WHERE users.id=work_items.reviewer_id),reviewer)::text AS reviewer, reference, event_codes, qa_cycle, return_count, created_at, updated_at, completed_at, COALESCE(package_id::text,'')::text AS package_id, product, acceptance_criteria, due_date, COALESCE(release_id::text,'')::text AS release_id, COALESCE((SELECT name FROM releases WHERE releases.id=work_items.release_id),'')::text AS release_name, COALESCE(assignee_id::text,'')::text AS assignee_id, COALESCE(reviewer_id::text,'')::text AS reviewer_id, COALESCE((SELECT name FROM workstreams WHERE workstreams.code=work_items.kind),'')::text AS kind_name FROM work_items WHERE (sqlc.arg(kind)::text='' OR kind=sqlc.arg(kind)::text) AND (sqlc.arg(status)::text='' OR status=sqlc.arg(status)::text) AND (sqlc.arg(search)::text='' OR title ILIKE '%'||sqlc.arg(search)::text||'%' OR reference ILIKE '%'||sqlc.arg(search)::text||'%') ORDER BY created_at DESC,number DESC;

-- name: GetDemand :one
SELECT id, number, title, description, kind, priority, status, COALESCE((SELECT name FROM users WHERE users.id=work_items.assignee_id),assignee)::text AS assignee, COALESCE((SELECT name FROM users WHERE users.id=work_items.reviewer_id),reviewer)::text AS reviewer, reference, event_codes, qa_cycle, return_count, created_at, updated_at, completed_at, COALESCE(package_id::text,'')::text AS package_id, product, acceptance_criteria, due_date, COALESCE(release_id::text,'')::text AS release_id, COALESCE((SELECT name FROM releases WHERE releases.id=work_items.release_id),'')::text AS release_name, COALESCE(assignee_id::text,'')::text AS assignee_id, COALESCE(reviewer_id::text,'')::text AS reviewer_id, COALESCE((SELECT name FROM workstreams WHERE workstreams.code=work_items.kind),'')::text AS kind_name FROM work_items WHERE work_items.id=sqlc.arg(id);

-- name: ListDemandTests :many
SELECT id,demand_id,cycle,title,steps,expected,actual,result,author,created_at,COALESCE(author_id::text,'')::text AS author_id FROM test_cases WHERE demand_id=sqlc.arg(demand_id) ORDER BY created_at ASC;

-- name: ListDemandTransitions :many
SELECT id,demand_id,from_status,to_status,actor,reason,created_at,COALESCE(actor_id::text,'')::text AS actor_id FROM work_item_transitions WHERE demand_id=sqlc.arg(demand_id) ORDER BY created_at ASC;

-- name: GetReleaseAssignmentForShare :one
SELECT name,archived_at IS NULL OR EXISTS(SELECT 1 FROM work_items WHERE id=sqlc.arg(demand_id)::uuid AND release_id=releases.id) AS allowed FROM releases WHERE id=sqlc.arg(release_id)::text::uuid FOR SHARE;

-- name: GetBoardColumns :many
WITH filtered AS (
 SELECT w.id,w.number,w.title,left(w.description,220) AS description,w.kind,ws.name AS kind_name,w.priority,w.status,COALESCE(a.name,w.assignee) AS assignee,COALESCE(q.name,w.reviewer) AS reviewer,COALESCE(w.assignee_id::text,'')::text AS assignee_id,COALESCE(w.reviewer_id::text,'')::text AS reviewer_id,w.reference,w.event_codes,w.qa_cycle,w.return_count,w.created_at,w.updated_at,w.completed_at,w.package_id,w.product,w.due_date,w.release_id,COALESCE(r.name,'') AS release_name FROM work_items w JOIN workstreams ws ON ws.code=w.kind LEFT JOIN releases r ON r.id=w.release_id LEFT JOIN users a ON a.id=w.assignee_id LEFT JOIN users q ON q.id=w.reviewer_id
 WHERE (sqlc.arg(release_id)::text='' OR (sqlc.arg(release_id)::text='none' AND w.release_id IS NULL) OR w.release_id::text=sqlc.arg(release_id)::text)
 AND (sqlc.arg(archive)::text='all' OR (sqlc.arg(archive)::text='active' AND r.archived_at IS NULL) OR (sqlc.arg(archive)::text='archived' AND r.archived_at IS NOT NULL))
 AND (sqlc.arg(kind)::text='' OR w.kind=sqlc.arg(kind)::text) AND (sqlc.arg(owner)::text='' OR w.assignee_id::text=sqlc.arg(owner)::text)
 AND (sqlc.arg(search)::text='' OR concat_ws(' ',w.title,w.reference,w.product,w.description,array_to_string(w.event_codes,' '),'SF-'||lpad(w.number::text,greatest(4,length(w.number::text)),'0')) ILIKE '%'||sqlc.arg(search)::text||'%')
 AND (sqlc.arg(status)::text='' OR w.status=sqlc.arg(status)::text)

), totals AS (
 SELECT f.status,count(*)::int AS total FROM filtered f GROUP BY f.status
)
SELECT totals.status,totals.total,COALESCE((
 SELECT jsonb_agg(to_jsonb(page) ORDER BY page.updated_at DESC,page.number DESC)
 FROM (SELECT f.* FROM filtered f WHERE f.status=totals.status
 ORDER BY f.updated_at DESC,f.number DESC LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int) page
),'[]'::jsonb)::jsonb AS items FROM totals;

-- name: ListBoardOwners :many
SELECT DISTINCT assignee_id::text FROM work_items WHERE assignee_id IS NOT NULL ORDER BY assignee_id::text;
