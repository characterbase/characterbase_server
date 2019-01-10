CREATE TABLE universes (
    id text PRIMARY KEY,
    name text NOT NULL,
    description text,
    guide jsonb NOT NULL,
    settings jsonb NOT NULL
);