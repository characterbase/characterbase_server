CREATE TABLE characters (
    id text PRIMARY KEY,
    universe_id text REFERENCES universes(id),
    owner_id text REFERENCES users(id),
    name text NOT NULL,
    tag text,
    fields jsonb NOT NULL,
    meta jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    UNIQUE (universe_id, name, tag)
);

CREATE INDEX universe_idx ON characters(universe_id);
CREATE INDEX name_idx ON characters(universe_id, name);
CREATE INDEX owner_idx ON characters(universe_id, owner_id);