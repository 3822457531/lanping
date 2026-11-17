-- +goose Up
ALTER TABLE lp_project
  MODIFY COLUMN card_mode VARCHAR(16) NOT NULL DEFAULT 'amount' COMMENT 'times次卡 amount金额 both兼容';

-- 仅空值改为金额；已有 both/times 保持原样，改为金额时由业务侧做次卡转余额
UPDATE lp_project SET card_mode='amount' WHERE card_mode='' OR card_mode IS NULL;

-- +goose Down
ALTER TABLE lp_project
  MODIFY COLUMN card_mode VARCHAR(16) NOT NULL DEFAULT 'both' COMMENT 'times次卡 amount金额 both兼容';
