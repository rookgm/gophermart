CREATE TABLE IF NOT EXISTS "orders" (
    "id" BIGSERIAL PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "number" varchar NOT NULL UNIQUE,
    "status" varchar NOT NULL,
    "accrual" numeric(10, 2),
    "uploaded_at" timestamptz NOT NULL DEFAULT (now()),
    FOREIGN KEY (user_id) REFERENCES users(id)
);