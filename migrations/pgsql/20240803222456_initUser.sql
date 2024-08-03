-- +goose Up
-- create "users" table
CREATE TABLE "users" (
  "id" bigserial NOT NULL,
  "login" text NOT NULL,
  "password" bytea NOT NULL,
  "balance" bigint NOT NULL DEFAULT 0,
  "withdrawn" bigint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_login" to table: "users"
CREATE UNIQUE INDEX "idx_login" ON "users" ("login");

-- +goose Down
-- reverse: create index "idx_login" to table: "users"
DROP INDEX "idx_login";
-- reverse: create "users" table
DROP TABLE "users";
