-- +goose Up
ALTER TABLE lp_customer_card
  MODIFY COLUMN project_id BIGINT NULL,
  ADD COLUMN card_kind VARCHAR(16) NOT NULL DEFAULT 'times' COMMENT 'times次卡/amount储值' AFTER pay_method;

UPDATE lp_customer_card SET card_kind='times' WHERE card_kind='' OR card_kind IS NULL;

-- +goose Down
ALTER TABLE lp_customer_card DROP COLUMN card_kind;
