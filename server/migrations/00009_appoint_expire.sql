-- +goose Up
INSERT INTO lp_settings (config_key, config_value) VALUES
('appoint_expire_minutes', '30')
ON DUPLICATE KEY UPDATE config_key=config_key;

-- +goose Down
DELETE FROM lp_settings WHERE config_key='appoint_expire_minutes';
