-- name: ConsumeUserResetTokensAt :exec
UPDATE password_reset_tokens SET consumed_at=sqlc.arg(now) WHERE user_id=sqlc.arg(user_id) AND consumed_at IS NULL;

-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens(token_hash,user_id,created_by,created_at,expires_at) VALUES(sqlc.arg(token_hash),sqlc.arg(user_id),sqlc.arg(actor_id),sqlc.arg(created_at),sqlc.arg(expires_at));

-- name: CreateResetLinkEvent :exec
INSERT INTO user_password_events(user_id,actor_id,action,created_at) VALUES(sqlc.arg(user_id),sqlc.arg(actor_id),'reset_link_created',sqlc.arg(created_at));

-- name: GetPasswordResetUser :one
SELECT user_id FROM password_reset_tokens WHERE token_hash=sqlc.arg(token_hash) AND consumed_at IS NULL AND expires_at>sqlc.arg(now) AND expires_at>clock_timestamp();

-- name: ConsumePasswordResetToken :execrows
UPDATE password_reset_tokens SET consumed_at=sqlc.arg(now) WHERE token_hash=sqlc.arg(token_hash) AND user_id=sqlc.arg(user_id) AND consumed_at IS NULL AND expires_at>sqlc.arg(now) AND expires_at>clock_timestamp();

-- name: GetPasswordForUpdate :one
SELECT password_hash FROM users WHERE id=sqlc.arg(user_id) FOR UPDATE;

-- name: UpdatePassword :exec
UPDATE users SET password_hash=sqlc.arg(password_hash) WHERE id=sqlc.arg(id);

-- name: RevokeUserSessions :exec
UPDATE auth_sessions SET revoked=true WHERE user_id=sqlc.arg(user_id) AND NOT revoked;

-- name: ConsumeUserResetTokens :exec
UPDATE password_reset_tokens SET consumed_at=clock_timestamp() WHERE user_id=sqlc.arg(user_id) AND consumed_at IS NULL;

-- name: CreatePasswordEvent :exec
INSERT INTO user_password_events(user_id,actor_id,action,created_at) VALUES(sqlc.arg(user_id),sqlc.arg(actor_id),sqlc.arg(action),clock_timestamp());
