CREATE TABLE password_reset_tokens (
 token_hash text PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 created_by uuid NOT NULL REFERENCES users(id),
 created_at timestamptz NOT NULL,
 expires_at timestamptz NOT NULL,
 consumed_at timestamptz,
 CHECK (expires_at > created_at)
);
CREATE INDEX password_reset_tokens_user ON password_reset_tokens(user_id);
CREATE TABLE user_password_events (
 id bigserial PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 actor_id uuid NOT NULL REFERENCES users(id),
 action text NOT NULL CHECK (action IN ('password_changed','reset_link_created','password_reset')),
 created_at timestamptz NOT NULL
);
CREATE INDEX user_password_events_user_date ON user_password_events(user_id,created_at DESC);
