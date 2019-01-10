CREATE TABLE users (
    id text PRIMARY KEY,
    email text UNIQUE NOT NULL,
    display_name text,
    password_hash text NOT NULL
);