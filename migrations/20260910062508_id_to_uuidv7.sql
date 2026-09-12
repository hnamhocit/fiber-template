-- Modify "users" table
ALTER TABLE "public"."users" ALTER COLUMN "id" TYPE uuid, ALTER COLUMN "id" SET DEFAULT uuidv7(), ALTER COLUMN "id" DROP IDENTITY;
