CREATE TABLE character_images (
    character_id text REFERENCES characters(id) ON DELETE CASCADE,
    key text NOT NULL,
    url text NOT NULL,
    PRIMARY KEY (character_id, key)
);