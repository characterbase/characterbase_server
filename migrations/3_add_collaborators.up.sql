CREATE TABLE collaborators (
    universe_id text REFERENCES universes(id),
    user_id text REFERENCES users(id),
    role integer,
    PRIMARY KEY (universe_id, user_id)
);