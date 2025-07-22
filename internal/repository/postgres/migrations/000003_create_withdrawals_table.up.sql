CREATE TABLE IF NOT EXISTS "withdrawals" (
    "id" BIGSERIAL PRIMARY KEY,
    "user_id" bigint NOT NULL,
    "order_number" varchar NOT NULL UNIQUE,
    "amount" numeric(10, 2) NOT NULL,
    "processed_at" timestamptz NOT NULL DEFAULT (now()),
    FOREIGN KEY (user_id) REFERENCES users(id)
    );