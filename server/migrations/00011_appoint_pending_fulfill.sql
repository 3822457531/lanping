-- +goose Up
-- 新建=待履约（非已履约）；到店员工操作后才=已履约
UPDATE lp_appointment SET status='待履约' WHERE status IN ('待确认','已确认');
-- 此前误把「排期确认」写成已履约、但时间未到的，改回待履约
UPDATE lp_appointment SET status='待履约'
WHERE status='已履约'
  AND TIMESTAMP(appoint_date, appoint_time) >= NOW();

-- +goose Down
UPDATE lp_appointment SET status='待确认' WHERE status='待履约';
