CREATE TABLE workstreams (
 id uuid PRIMARY KEY,
 code varchar(64) NOT NULL UNIQUE,
 name varchar(80) NOT NULL CHECK (length(trim(name)) BETWEEN 2 AND 80),
 color varchar(7) NOT NULL CHECK (color ~ '^#[0-9a-fA-F]{6}$'),
 active boolean NOT NULL DEFAULT true,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX workstreams_name ON workstreams(lower(trim(name)));
-- Preserve the existing categories and their references on upgrades.
INSERT INTO workstreams(id,code,name,color) VALUES
 ('66666666-6666-4666-8666-666666666601','improvement','Melhorias','#8976e8'),
 ('66666666-6666-4666-8666-666666666602','corrective','Corretivas','#e77c99'),
 ('66666666-6666-4666-8666-666666666603','testing','Testes','#57bad3'),
 ('66666666-6666-4666-8666-666666666604','rewrite','Reescritas','#e9ad67');
ALTER TABLE work_items DROP CONSTRAINT work_items_kind_check;
ALTER TABLE work_packages DROP CONSTRAINT work_packages_kind_check;
ALTER TABLE work_items ALTER COLUMN kind TYPE varchar(64);
ALTER TABLE work_packages ALTER COLUMN kind TYPE varchar(64);
ALTER TABLE work_items ADD CONSTRAINT work_items_kind_fkey FOREIGN KEY(kind) REFERENCES workstreams(code);
ALTER TABLE work_packages ADD CONSTRAINT work_packages_kind_fkey FOREIGN KEY(kind) REFERENCES workstreams(code);
ALTER TABLE work_items ADD COLUMN assignee_id uuid REFERENCES users(id);
ALTER TABLE work_items ADD COLUMN reviewer_id uuid REFERENCES users(id);
CREATE INDEX work_items_assignee_id ON work_items(assignee_id);
CREATE INDEX work_items_reviewer_id ON work_items(reviewer_id);
-- Match only unambiguous registered identities. Unmatched text remains historical.
WITH unique_names AS (SELECT lower(trim(name)) AS name, min(id::text)::uuid AS id FROM users GROUP BY lower(trim(name)) HAVING count(*)=1)
UPDATE work_items w SET assignee_id=u.id FROM unique_names u WHERE lower(trim(w.assignee))=u.name;
WITH unique_names AS (SELECT lower(trim(name)) AS name, min(id::text)::uuid AS id FROM users GROUP BY lower(trim(name)) HAVING count(*)=1)
UPDATE work_items w SET reviewer_id=u.id FROM unique_names u WHERE lower(trim(w.reviewer))=u.name;
ALTER TABLE work_item_transitions ADD COLUMN actor_id uuid REFERENCES users(id);
ALTER TABLE test_cases ADD COLUMN author_id uuid REFERENCES users(id);
UPDATE work_item_transitions t SET actor_id=u.id FROM users u WHERE right(t.actor,38)='('||u.id::text||')';
UPDATE test_cases t SET author_id=u.id FROM users u WHERE right(t.author,38)='('||u.id::text||')';
