-- +goose Up
INSERT INTO lp_settings (config_key, config_value) VALUES
('verify_mode', 'mock'),
('customer_share', 'false')
ON DUPLICATE KEY UPDATE config_key=config_key;

-- +goose Down
DELETE FROM lp_settings WHERE config_key IN ('verify_mode','customer_share');
