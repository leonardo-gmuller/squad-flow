-- name: ListNotifications :one
WITH page AS (
 SELECT n.id,n.demand_id,n.kind,n.title,n.message,n.read_at,n.created_at
 FROM notifications n WHERE n.user_id=sqlc.arg(user_id)
 AND (NOT sqlc.arg(unread)::boolean OR n.read_at IS NULL)
 AND (sqlc.narg(before)::timestamptz IS NULL OR (n.created_at,n.id)<(sqlc.narg(before),sqlc.narg(before_id)::uuid))
 ORDER BY n.created_at DESC,n.id DESC LIMIT sqlc.arg(page_limit)::int
)
SELECT COALESCE(jsonb_agg(to_jsonb(page) ORDER BY page.created_at DESC,page.id DESC),'[]'::jsonb)::jsonb AS items,
 (SELECT count(*)::int FROM notifications n WHERE n.user_id=sqlc.arg(user_id) AND n.read_at IS NULL) AS unread_count
FROM page;

-- name: UnreadNotificationCount :one
SELECT count(*)::int FROM notifications WHERE notifications.user_id=sqlc.arg(user_id) AND read_at IS NULL;

-- name: ReadNotification :one
UPDATE notifications SET read_at=COALESCE(read_at,now()) WHERE id=sqlc.arg(id) AND user_id=sqlc.arg(user_id) RETURNING COALESCE(read_at,now())::timestamptz AS read_at;

-- name: ReadAllNotifications :one
WITH updated AS (UPDATE notifications SET read_at=now() WHERE notifications.user_id=sqlc.arg(user_id) AND read_at IS NULL) SELECT now()::timestamptz AS read_at;

-- name: CreateNotification :exec
INSERT INTO notifications(id,user_id,demand_id,kind,title,message)
 SELECT sqlc.arg(id)::uuid,users.id,sqlc.arg(demand_id),sqlc.arg(kind),sqlc.arg(title),sqlc.arg(message) FROM users WHERE users.id=sqlc.arg(user_id) AND role='member';

-- name: GetDemandNotificationRecipients :one
SELECT id,number,title,COALESCE(assignee_id::text,'')::text AS assignee_id,COALESCE(reviewer_id::text,'')::text AS reviewer_id FROM work_items WHERE id=sqlc.arg(id);
