-- +goose Up
ALTER TABLE "entities" ADD COLUMN IF NOT EXISTS "minimum_quantity" double precision;
ALTER TABLE "entities" ADD COLUMN IF NOT EXISTS "maximum_quantity" double precision;
ALTER TABLE "entities" ADD COLUMN IF NOT EXISTS "reorder_threshold" double precision;

-- +goose Down
ALTER TABLE "entities" DROP COLUMN IF EXISTS "reorder_threshold";
ALTER TABLE "entities" DROP COLUMN IF EXISTS "maximum_quantity";
ALTER TABLE "entities" DROP COLUMN IF EXISTS "minimum_quantity";
