-- +goose Up
ALTER TABLE lp_welfare
  ADD COLUMN valid_days INT NOT NULL DEFAULT 30 COMMENT '发放后有效天数' AFTER amount;

ALTER TABLE lp_welfare_issued
  ADD COLUMN remain_amount DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '剩余可抵金额' AFTER amount,
  ADD COLUMN used_amount DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '已抵金额' AFTER remain_amount,
  ADD COLUMN expire_at DATETIME NULL COMMENT '截止有效时间' AFTER used_amount,
  ADD COLUMN status VARCHAR(16) NOT NULL DEFAULT '可用' COMMENT '可用/已用完/已过期' AFTER expire_at;

UPDATE lp_welfare_issued
SET remain_amount = amount,
    used_amount = 0,
    expire_at = DATE_ADD(issued_at, INTERVAL 30 DAY),
    status = '可用'
WHERE remain_amount = 0 AND used_amount = 0 AND (expire_at IS NULL OR status = '' OR status IS NULL);

ALTER TABLE lp_verify_order
  ADD COLUMN welfare_issued_id BIGINT NULL COMMENT '抵扣的福利发放记录' AFTER times,
  ADD COLUMN welfare_offset DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '福利抵扣金额' AFTER welfare_issued_id,
  ADD COLUMN pay_amount DECIMAL(12,2) NOT NULL DEFAULT 0 COMMENT '实际从卡扣减金额' AFTER welfare_offset;

UPDATE lp_verify_order SET pay_amount = total_amount WHERE pay_amount = 0;

-- +goose Down
ALTER TABLE lp_verify_order
  DROP COLUMN welfare_issued_id,
  DROP COLUMN welfare_offset,
  DROP COLUMN pay_amount;

ALTER TABLE lp_welfare_issued
  DROP COLUMN remain_amount,
  DROP COLUMN used_amount,
  DROP COLUMN expire_at,
  DROP COLUMN status;

ALTER TABLE lp_welfare DROP COLUMN valid_days;
