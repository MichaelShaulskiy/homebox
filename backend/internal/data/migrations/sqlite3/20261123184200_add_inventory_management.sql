-- +goose Up
ALTER TABLE entities ADD COLUMN available_quantity real NOT NULL DEFAULT 1;
ALTER TABLE entities ADD COLUMN defective_quantity real NOT NULL DEFAULT 0;
UPDATE entities SET available_quantity = quantity;

CREATE TABLE IF NOT EXISTS inventory_recipients (
    id uuid NOT NULL PRIMARY KEY,
    created_at datetime NOT NULL,
    updated_at datetime NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    email text,
    department text,
    is_active boolean NOT NULL DEFAULT 1,
    group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS inventory_recipient_group_email
    ON inventory_recipients(group_id, email);
CREATE INDEX IF NOT EXISTS inventory_recipient_group_active
    ON inventory_recipients(group_id, is_active);

CREATE TABLE IF NOT EXISTS inventory_assignments (
    id uuid NOT NULL PRIMARY KEY,
    created_at datetime NOT NULL,
    updated_at datetime NOT NULL,
    quantity real NOT NULL,
    serial_number text,
    asset_tag text,
    note text,
    issued_at datetime NOT NULL,
    returned_at datetime,
    group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    entity_id uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    recipient_id uuid NOT NULL REFERENCES inventory_recipients(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS inventory_assignment_group
    ON inventory_assignments(group_id);
CREATE INDEX IF NOT EXISTS inventory_assignment_active
    ON inventory_assignments(group_id, returned_at);

CREATE TABLE IF NOT EXISTS inventory_transactions (
    id uuid NOT NULL PRIMARY KEY,
    created_at datetime NOT NULL,
    updated_at datetime NOT NULL,
    timestamp datetime NOT NULL,
    action_type text NOT NULL CHECK (action_type IN (
        'ISSUED', 'RETURNED_OK', 'RETURNED_DEFECTIVE', 'WRITTEN_OFF', 'STOCK_ADDED'
    )),
    item_id uuid NOT NULL,
    item_name text NOT NULL,
    quantity real NOT NULL,
    performed_by_admin_id uuid NOT NULL,
    performed_by_admin_name text NOT NULL,
    target_user_id uuid,
    target_user_name text,
    reason_or_note text,
    serial_number text,
    asset_tag text,
    group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS inventory_transaction_group
    ON inventory_transactions(group_id);
CREATE INDEX IF NOT EXISTS inventory_transaction_timestamp
    ON inventory_transactions(timestamp);
CREATE INDEX IF NOT EXISTS inventory_transaction_action
    ON inventory_transactions(group_id, action_type);
CREATE INDEX IF NOT EXISTS inventory_transaction_item
    ON inventory_transactions(item_id);
CREATE INDEX IF NOT EXISTS inventory_transaction_user
    ON inventory_transactions(target_user_id);
CREATE INDEX IF NOT EXISTS inventory_transaction_admin
    ON inventory_transactions(performed_by_admin_id);

-- +goose Down
DROP TABLE IF EXISTS inventory_transactions;
DROP TABLE IF EXISTS inventory_assignments;
DROP TABLE IF EXISTS inventory_recipients;
ALTER TABLE entities DROP COLUMN defective_quantity;
ALTER TABLE entities DROP COLUMN available_quantity;
