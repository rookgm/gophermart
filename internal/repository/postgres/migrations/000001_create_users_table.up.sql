CREATE TABLE IF NOT EXISTS "users" (
    "id" BIGSERIAL PRIMARY KEY,
    "login" varchar NOT NULL UNIQUE,
    "password" varchar NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);