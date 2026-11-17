-- +goose Up
-- 预约状态约定：待确认 / 已履约 / 已改约 / 已取消 / 已失效
UPDATE lp_appointment SET status='已履约' WHERE status='已确认';
UPDATE lp_appointment SET status='已失效' WHERE status IN ('失效','已过期');
-- 旧「已完成」视为历史履约完成，归入已履约后由业务侧不再占用：改为已失效不合适；
-- 已服务结束的旧单进入历史，用已履约且日期已过会在列表进历史筛选——见前端：已完成兼容显示。
-- 此处将旧已完成保留字段语义迁移为已履约（终态展示由前端 history 含已完成兼容，后端占用仅认三种进行中）
UPDATE lp_appointment SET status='已履约' WHERE status='已完成';

UPDATE lp_settings SET config_value='超过预约开始时间未确认/改约/取消则自动「已失效」'
WHERE config_key='appoint_expire_hint' AND 1=0;

-- +goose Down
UPDATE lp_appointment SET status='已确认' WHERE status='已履约';
UPDATE lp_appointment SET status='失效' WHERE status='已失效';
