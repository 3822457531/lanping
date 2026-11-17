-- +goose Up
ALTER TABLE lp_welfare_issued
  MODIFY COLUMN user_id BIGINT NULL COMMENT '可选：指定客户，空表示按门店发放';

-- +goose Down
UPDATE lp_welfare_issued SET user_id=0 WHERE user_id IS NULL;
ALTER TABLE lp_welfare_issued
  MODIFY COLUMN user_id BIGINT NOT NULL;
