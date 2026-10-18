CREATE TABLE releases (
    id uuid PRIMARY KEY,
    name varchar(80) NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 80),
    description text NOT NULL DEFAULT '',
    release_date timestamptz,
    archived_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
CREATE UNIQUE INDEX idx_releases_name ON releases (lower(trim(name)));
ALTER TABLE work_items ADD COLUMN release_id uuid REFERENCES releases(id);
CREATE INDEX idx_work_items_release_status ON work_items(release_id, status, updated_at DESC, number DESC);
CREATE INDEX idx_work_items_board ON work_items(status, updated_at DESC, number DESC);
