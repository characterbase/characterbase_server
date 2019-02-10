CREATE TABLE collaborators (
    universe_id text REFERENCES universes(id) ON DELETE CASCADE,
    user_id text REFERENCES users(id) ON DELETE CASCADE,
    role integer,
    PRIMARY KEY (universe_id, user_id)
);