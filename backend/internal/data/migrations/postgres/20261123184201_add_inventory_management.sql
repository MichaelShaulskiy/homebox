-- +goose Up
ALTER TABLE "entities" ADD COLUMN IF NOT EXISTS "available_quantity" double precision NOT NULL DEFAULT 1;
ALTER TABLE "entities" ADD COLUMN IF NOT EXISTS "defective_quantity" double precision NOT NULL DEFAULT 0;
UPDATE "entities" SET "available_quantity" = "quantity";

CREATE TABLE IF NOT EXISTS "inventory_recipients" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "first_name" character varying NOT NULL,
    "last_name" character varying NOT NULL,
    "email" character varying,
    "department" character varying,
    "is_active" boolean NOT NULL DEFAULT true,
    "group_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "inventory_recipients_groups" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "inventory_recipient_group_email"
    ON "inventory_recipients" ("group_id", "email");
CREATE INDEX IF NOT EXISTS "inventory_recipient_group_active"
    ON "inventory_recipients" ("group_id", "is_active");

CREATE TABLE IF NOT EXISTS "inventory_assignments" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "quantity" double precision NOT NULL,
    "serial_number" character varying,
    "asset_tag" character varying,
    "note" character varying,
    "issued_at" timestamptz NOT NULL,
    "returned_at" timestamptz,
    "group_id" uuid NOT NULL,
    "entity_id" uuid NOT NULL,
    "recipient_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "inventory_assignments_groups" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE,
    CONSTRAINT "inventory_assignments_entities" FOREIGN KEY ("entity_id") REFERENCES "entities" ("id") ON DELETE CASCADE,
    CONSTRAINT "inventory_assignments_recipients" FOREIGN KEY ("recipient_id") REFERENCES "inventory_recipients" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "inventory_assignment_group"
    ON "inventory_assignments" ("group_id");
CREATE INDEX IF NOT EXISTS "inventory_assignment_active"
    ON "inventory_assignments" ("group_id", "returned_at");

CREATE TABLE IF NOT EXISTS "inventory_transactions" (
    "id" uuid NOT NULL,
    "created_at" timestamptz NOT NULL,
    "updated_at" timestamptz NOT NULL,
    "timestamp" timestamptz NOT NULL,
    "action_type" character varying NOT NULL CHECK ("action_type" IN (
        'ISSUED', 'RETURNED_OK', 'RETURNED_DEFECTIVE', 'WRITTEN_OFF', 'STOCK_ADDED'
    )),
    "item_id" uuid NOT NULL,
    "item_name" character varying NOT NULL,
    "quantity" double precision NOT NULL,
    "performed_by_admin_id" uuid NOT NULL,
    "performed_by_admin_name" text NOT NULL,
    "target_user_id" uuid,
    "target_user_name" character varying,
    "reason_or_note" character varying,
    "serial_number" character varying,
    "asset_tag" character varying,
    "group_id" uuid NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "inventory_transactions_groups" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "inventory_transaction_group"
    ON "inventory_transactions" ("group_id");
CREATE INDEX IF NOT EXISTS "inventory_transaction_timestamp"
    ON "inventory_transactions" ("timestamp");
CREATE INDEX IF NOT EXISTS "inventory_transaction_action"
    ON "inventory_transactions" ("group_id", "action_type");
CREATE INDEX IF NOT EXISTS "inventory_transaction_item"
    ON "inventory_transactions" ("item_id");
CREATE INDEX IF NOT EXISTS "inventory_transaction_user"
    ON "inventory_transactions" ("target_user_id");
CREATE INDEX IF NOT EXISTS "inventory_transaction_admin"
    ON "inventory_transactions" ("performed_by_admin_id");

-- +goose Down
DROP TABLE IF EXISTS "inventory_transactions";
DROP TABLE IF EXISTS "inventory_assignments";
DROP TABLE IF EXISTS "inventory_recipients";
ALTER TABLE "entities" DROP COLUMN IF EXISTS "defective_quantity";
ALTER TABLE "entities" DROP COLUMN IF EXISTS "available_quantity";
