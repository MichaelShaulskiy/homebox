-- +goose Up
ALTER TABLE entities ADD COLUMN minimum_quantity real;
ALTER TABLE entities ADD COLUMN maximum_quantity real;
ALTER TABLE entities ADD COLUMN reorder_threshold real;

-- +goose Down
ALTER TABLE entities DROP COLUMN reorder_threshold;
ALTER TABLE entities DROP COLUMN maximum_quantity;
ALTER TABLE entities DROP COLUMN minimum_quantity;
