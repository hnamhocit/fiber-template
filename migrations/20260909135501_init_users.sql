-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigint NOT NULL GENERATED ALWAYS AS IDENTITY,
  "email" text NOT NULL,
  "name" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "users_email_key" UNIQUE ("email")
);
