-- name: LockAdminBootstrap :exec
SELECT pg_advisory_xact_lock(77621892);

-- name: AdminExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE role='admin');

-- name: CreateAdmin :exec
INSERT INTO users(id,name,email,role,password_hash,created_at) VALUES(sqlc.arg(id),sqlc.arg(name),sqlc.arg(email),'admin',sqlc.arg(password_hash),sqlc.arg(created_at));

-- name: GetUserByEmail :one
SELECT u.id,u.name,u.email,u.role,u.password_hash,u.created_at FROM users u WHERE email=sqlc.arg(email);

-- name: CreateUser :exec
INSERT INTO users(id,name,email,role,password_hash,created_at) VALUES(sqlc.arg(id),sqlc.arg(name),sqlc.arg(email),'member',sqlc.arg(password_hash),sqlc.arg(created_at));

-- name: ListUsers :many
SELECT u.id,u.name,u.email,u.role,u.password_hash,u.created_at FROM users u ORDER BY created_at;

-- name: LockUserPassword :one
SELECT password_hash FROM users WHERE id=sqlc.arg(id) FOR UPDATE;

-- name: CreateSession :exec
INSERT INTO auth_sessions(id,user_id,expires_at) VALUES(sqlc.arg(id),sqlc.arg(user_id),sqlc.arg(expires_at));

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens(token_hash,session_id) VALUES(sqlc.arg(token_hash),sqlc.arg(session_id));

-- name: LockRefreshSession :one
SELECT s.id,s.user_id,s.expires_at,t.consumed,s.revoked FROM auth_sessions s JOIN refresh_tokens t ON t.session_id=s.id WHERE t.token_hash=sqlc.arg(token_hash) FOR UPDATE OF s,t;

-- name: RevokeSession :exec
UPDATE auth_sessions SET revoked=true WHERE id=sqlc.arg(id);

-- name: ConsumeRefreshToken :exec
UPDATE refresh_tokens SET consumed=true WHERE token_hash=sqlc.arg(token_hash);

-- name: RotateRefreshToken :exec
INSERT INTO refresh_tokens(token_hash,session_id) VALUES(sqlc.arg(token_hash),sqlc.arg(session_id));

-- name: GetSessionUser :one
SELECT u.id,u.name,u.email,u.role,u.password_hash,u.created_at FROM users u JOIN auth_sessions s ON s.user_id=u.id WHERE s.id=sqlc.arg(id) AND NOT s.revoked AND s.expires_at>sqlc.arg(now);

-- name: RevokeRefreshSession :exec
UPDATE auth_sessions SET revoked=true WHERE id IN(SELECT session_id FROM refresh_tokens WHERE token_hash=sqlc.arg(token_hash));
