-- +goose Up
ALTER TABLE lp_stock_order
  ADD COLUMN to_store_id BIGINT NULL COMMENT '门店调度目标门店' AFTER store_id,
  ADD COLUMN confirm_by BIGINT NULL COMMENT '目标店确认人' AFTER operator_id,
  ADD COLUMN confirm_at DATETIME NULL COMMENT '确认时间' AFTER confirm_by;

-- +goose Down
ALTER TABLE lp_stock_order
  DROP COLUMN to_store_id,
  DROP COLUMN confirm_by,
  DROP COLUMN confirm_at;
