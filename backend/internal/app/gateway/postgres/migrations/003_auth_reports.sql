CREATE TABLE users (
 id uuid PRIMARY KEY,
 name varchar(120) NOT NULL,
 email varchar(254) NOT NULL UNIQUE CHECK(email = lower(email)),
 password_hash text NOT NULL,
 role text NOT NULL CHECK(role IN ('admin','member')),
 created_at timestamptz NOT NULL
);
CREATE UNIQUE INDEX users_single_admin ON users(role) WHERE role = 'admin';
CREATE TABLE auth_sessions (
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id),
 expires_at timestamptz NOT NULL,
 revoked boolean NOT NULL DEFAULT false
);
CREATE INDEX auth_sessions_user ON auth_sessions(user_id);
CREATE TABLE refresh_tokens (
 token_hash text PRIMARY KEY,
 session_id uuid NOT NULL REFERENCES auth_sessions(id) ON DELETE CASCADE,
 consumed boolean NOT NULL DEFAULT false
);
CREATE INDEX refresh_tokens_session ON refresh_tokens(session_id);
CREATE TABLE generated_reports (
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id),
 demand_id uuid REFERENCES work_items(id),
 prompt text NOT NULL,
 payload jsonb NOT NULL,
 created_at timestamptz NOT NULL
);
CREATE INDEX generated_reports_user_date ON generated_reports(user_id,created_at DESC);
