CREATE TABLE notifications (
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 demand_id uuid NOT NULL REFERENCES work_items(id) ON DELETE CASCADE,
 kind varchar(32) NOT NULL CHECK(kind IN ('assigned','unassigned','status_changed','qa_passed')),
 title varchar(160) NOT NULL,
 message text NOT NULL,
 read_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notifications_user_page ON notifications(user_id,created_at DESC,id DESC);
CREATE INDEX notifications_user_unread ON notifications(user_id) WHERE read_at IS NULL;
