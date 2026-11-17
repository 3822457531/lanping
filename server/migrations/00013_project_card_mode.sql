-- +goose Up
ALTER TABLE lp_project
  ADD COLUMN card_mode VARCHAR(16) NOT NULL DEFAULT 'both' COMMENT 'times次卡 amount金额 both兼容' AFTER total_times;

UPDATE lp_project SET card_mode='both' WHERE card_mode='' OR card_mode IS NULL;

-- +goose Down
ALTER TABLE lp_project DROP COLUMN card_mode;
