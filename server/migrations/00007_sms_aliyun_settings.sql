-- +goose Up
INSERT INTO lp_settings (config_key, config_value) VALUES
('sms_access_key_id', ''),
('sms_access_key_secret', ''),
('sms_endpoint', 'dysmsapi.aliyuncs.com'),
('sms_tpl_birthday', ''),
('sms_tpl_consume', ''),
('sms_tpl_appoint', '')
ON DUPLICATE KEY UPDATE config_key=config_key;

UPDATE lp_settings SET config_value='兰瓶' WHERE config_key='sms_sign' AND (config_value='' OR config_value='兰瓶美业');

-- +goose Down
DELETE FROM lp_settings WHERE config_key IN (
  'sms_access_key_id','sms_access_key_secret','sms_endpoint',
  'sms_tpl_birthday','sms_tpl_consume','sms_tpl_appoint'
);
