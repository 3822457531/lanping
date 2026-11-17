-- +goose Up
ALTER TABLE lp_store
  ADD COLUMN lat DECIMAL(10,7) NOT NULL DEFAULT 0 COMMENT '打卡纬度' AFTER address,
  ADD COLUMN lng DECIMAL(10,7) NOT NULL DEFAULT 0 COMMENT '打卡经度' AFTER lat;

ALTER TABLE lp_attendance
  ADD COLUMN lat DECIMAL(10,7) NULL COMMENT '打卡纬度' AFTER method,
  ADD COLUMN lng DECIMAL(10,7) NULL COMMENT '打卡经度' AFTER lat,
  ADD COLUMN distance_m INT NULL COMMENT '距门店米数' AFTER lng;

-- +goose Down
ALTER TABLE lp_attendance
  DROP COLUMN lat,
  DROP COLUMN lng,
  DROP COLUMN distance_m;

ALTER TABLE lp_store
  DROP COLUMN lat,
  DROP COLUMN lng;
