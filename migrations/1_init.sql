-- +goose Up

CREATE TABLE "users" (
  "user_id" text PRIMARY KEY,
  "full_name" text,
  "email" text UNIQUE,
  "password" text,
  "verify" boolean,
  "create_at" TIMESTAMPTZ NOT NULL,
  "update_at" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "posts" (
  "link" text PRIMARY KEY,
  "name" text,
  "tag"  text
);

CREATE TABLE "bookmarks" (
  "bookmark_id" text PRIMARY KEY,
  "user_id" text,
  "post_name" text,
  "created_at" TIMESTAMPTZ NOT NULL,
  "updated_at" TIMESTAMPTZ NOT NULL,
  unique (user_id, post_name)
);

ALTER TABLE "bookmarks" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("user_id");
ALTER TABLE "bookmarks" ADD FOREIGN KEY ("post_name") REFERENCES "posts" ("link");

