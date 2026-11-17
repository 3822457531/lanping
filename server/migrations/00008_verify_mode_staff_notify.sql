-- +goose Up
-- mock → staff（不发通知，员工自助）；real → notify（发通知，客户亮码）
UPDATE lp_settings SET config_value='staff' WHERE config_key='verify_mode' AND config_value IN ('mock','');
UPDATE lp_settings SET config_value='notify' WHERE config_key='verify_mode' AND config_value='real';
INSERT INTO lp_settings (config_key, config_value) VALUES ('verify_mode', 'staff')
ON DUPLICATE KEY UPDATE config_key=config_key;

-- +goose Down
UPDATE lp_settings SET config_value='mock' WHERE config_key='verify_mode' AND config_value='staff';
UPDATE lp_settings SET config_value='real' WHERE config_key='verify_mode' AND config_value='notify';
