-- +goose Up
-- create "users" table
CREATE TABLE "users" (
  "id" bigserial NOT NULL,
  "login" character varying(32) NOT NULL,
  "password" bytea NOT NULL,
  "balance" bigint NOT NULL DEFAULT 0,
  "withdrawn" bigint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_login" to table: "users"
CREATE UNIQUE INDEX "idx_login" ON "users" ("login");
-- create "orders" table
CREATE TABLE "orders" (
  "id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "status" character varying(16) NOT NULL DEFAULT 'NEW',
  "accrual" bigint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_orders" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- create index "idx_created_at" to table: "orders"
CREATE INDEX "idx_created_at" ON "orders" ("created_at" DESC);
-- create "withdrawals" table
CREATE TABLE "withdrawals" (
  "order_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "sum" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("order_id"),
  CONSTRAINT "fk_users_withdrawals" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- +goose Down
-- reverse: create "withdrawals" table
DROP TABLE "withdrawals";
-- reverse: create index "idx_created_at" to table: "orders"
DROP INDEX "idx_created_at";
-- reverse: create "orders" table
DROP TABLE "orders";
-- reverse: create index "idx_login" to table: "users"
DROP INDEX "idx_login";
-- reverse: create "users" table
DROP TABLE "users";
