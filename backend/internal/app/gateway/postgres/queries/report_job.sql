-- name: LockReportUserQueue :exec
SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(user_id)::text, 77621893));

-- name: CreateReportJob :one
INSERT INTO report_jobs(id,user_id,request_key,request,max_attempts,trace_parent,trace_state)
VALUES(sqlc.arg(id),sqlc.arg(user_id),sqlc.arg(request_key),sqlc.arg(request),sqlc.arg(max_attempts),sqlc.arg(trace_parent),sqlc.arg(trace_state))
RETURNING *;

-- name: GetReportJobByRequestKey :one
SELECT * FROM report_jobs WHERE user_id=sqlc.arg(user_id) AND request_key=sqlc.arg(request_key);

-- name: GetReportJob :one
SELECT * FROM report_jobs WHERE id=sqlc.arg(id) AND user_id=sqlc.arg(user_id);

-- name: ListReportJobs :many
SELECT * FROM report_jobs WHERE user_id=sqlc.arg(user_id) ORDER BY created_at DESC,id DESC LIMIT 50;

-- name: CreateReportEvent :exec
INSERT INTO report_events(id,job_id,event_type,event_data,available_at,created_at,trace_parent,trace_state)
VALUES(sqlc.arg(id),sqlc.arg(job_id),sqlc.arg(event_type),sqlc.arg(event_data),sqlc.arg(available_at),clock_timestamp(),sqlc.arg(trace_parent),sqlc.arg(trace_state));

-- name: ListReportEvents :many
SELECT e.* FROM report_events e JOIN report_jobs j ON j.id=e.job_id
WHERE j.id=sqlc.arg(job_id) AND j.user_id=sqlc.arg(user_id)
ORDER BY e.created_at,e.id;

-- name: ClaimReportJob :one
UPDATE report_jobs SET status='processing',attempts=attempts+1,
    lease_token=sqlc.arg(lease_token)::uuid,lease_until=sqlc.arg(lease_until)::timestamptz,
    progress_agent='coleta',progress_message='Consultando os dados do workspace',error_message='',updated_at=now()
WHERE id=sqlc.arg(id) AND status='queued' AND available_at<=now() AND attempts<max_attempts
RETURNING *;

-- name: UpdateReportJobProgress :execrows
UPDATE report_jobs SET progress_agent=sqlc.arg(agent),progress_message=sqlc.arg(message),updated_at=now()
WHERE id=sqlc.arg(id) AND status='processing' AND lease_token=sqlc.arg(lease_token)::uuid AND lease_until>now();

-- name: CompleteReportJob :execrows
UPDATE report_jobs SET status='completed',report_id=sqlc.arg(report_id)::uuid,
    progress_agent='concluído',progress_message='Relatório pronto para visualizar e exportar',
    error_message='',lease_token=NULL,lease_until=NULL,updated_at=now(),completed_at=now()
WHERE id=sqlc.arg(id) AND status='processing' AND lease_token=sqlc.arg(lease_token)::uuid AND lease_until>now();

-- name: FailReportJob :one
UPDATE report_jobs SET status=CASE WHEN attempts<max_attempts THEN 'queued' ELSE 'failed' END,
    available_at=sqlc.arg(available_at),error_message=sqlc.arg(message),
    progress_agent=CASE WHEN attempts<max_attempts THEN 'fila' ELSE 'falha' END,
    progress_message=CASE WHEN attempts<max_attempts THEN 'Uma nova tentativa será feita automaticamente' ELSE 'Não foi possível concluir o relatório' END,
    lease_token=NULL,lease_until=NULL,updated_at=now()
WHERE id=sqlc.arg(id) AND status='processing' AND lease_token=sqlc.arg(lease_token)::uuid
RETURNING *;

-- name: ClaimReportOutbox :many
WITH pending AS (
    SELECT id FROM report_events WHERE event_type='report.requested' AND published_at IS NULL
        AND available_at<=now() AND (publish_lease_until IS NULL OR publish_lease_until<now())
    ORDER BY created_at,id FOR UPDATE SKIP LOCKED LIMIT sqlc.arg(page_limit)::int
)
UPDATE report_events e SET publish_token=sqlc.arg(publish_token)::uuid,publish_lease_until=sqlc.arg(lease_until)::timestamptz
FROM pending p WHERE e.id=p.id RETURNING e.*;

-- name: MarkReportEventPublished :execrows
UPDATE report_events SET published_at=now(),publish_lease_until=NULL,publish_token=NULL
WHERE id=sqlc.arg(id) AND publish_token=sqlc.arg(publish_token)::uuid;

-- name: ListRecoverableReportJobs :many
SELECT * FROM report_jobs WHERE
    (status='processing' AND lease_until<now()) OR
    (status='queued' AND updated_at<now()-interval '60 seconds' AND available_at<=now())
ORDER BY updated_at LIMIT 100;

-- name: RecoverReportJob :one
UPDATE report_jobs SET status=CASE WHEN attempts<max_attempts THEN 'queued' ELSE 'failed' END,
    lease_token=NULL,lease_until=NULL,available_at=now(),updated_at=now(),
    progress_agent=CASE WHEN attempts<max_attempts THEN 'fila' ELSE 'falha' END,
    progress_message=CASE WHEN attempts<max_attempts THEN 'Retomando o processamento do relatório' ELSE 'Processamento interrompido após o limite de tentativas' END
WHERE id=sqlc.arg(id) AND (
    (status='processing' AND lease_until<now()) OR
    (status='queued' AND updated_at<now()-interval '60 seconds' AND available_at<=now()))
RETURNING *;

-- name: LockReportJobLease :one
SELECT * FROM report_jobs WHERE id=sqlc.arg(id) AND status='processing'
AND lease_token=sqlc.arg(lease_token)::uuid AND lease_until>now() FOR UPDATE;
