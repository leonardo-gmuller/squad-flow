-- An OS / package groups zero or more independently-owned execution items.
-- Correction tickets without a package remain first-class work items.
CREATE TABLE work_packages (
 id uuid PRIMARY KEY,
 reference varchar(120) NOT NULL DEFAULT '',
 title varchar(180) NOT NULL,
 description text NOT NULL DEFAULT '',
 initiative varchar(160) NOT NULL DEFAULT '',
 kind varchar(24) NOT NULL CHECK(kind IN ('improvement','corrective','testing','rewrite')),
 created_at timestamptz NOT NULL,
 updated_at timestamptz NOT NULL
);
CREATE INDEX idx_packages_reference ON work_packages(reference);
CREATE INDEX idx_packages_initiative ON work_packages(initiative);
ALTER TABLE work_items ADD COLUMN package_id uuid REFERENCES work_packages(id) ON DELETE RESTRICT;
ALTER TABLE work_items ADD COLUMN product varchar(120) NOT NULL DEFAULT '';
ALTER TABLE work_items ADD COLUMN acceptance_criteria text NOT NULL DEFAULT '';
ALTER TABLE work_items ADD COLUMN due_date timestamptz;
CREATE INDEX idx_items_package ON work_items(package_id);
CREATE INDEX idx_items_assignee ON work_items(assignee);
