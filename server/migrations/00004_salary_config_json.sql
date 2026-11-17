-- +goose Up
ALTER TABLE lp_salary_rule
  ADD COLUMN config_json LONGTEXT NULL COMMENT '完整薪资规则JSON' AFTER late_deduct;

UPDATE lp_salary_rule SET name='兰瓶美业固定薪酬+开卡/核销提成', config_json='{"name":"兰瓶美业固定薪酬+开卡/核销提成","note":"销售提成=开卡提成；消耗提成=核销提成","manager":{"fixed_items":{"base_salary":1600,"overtime":700,"non_compete":100,"social_subsidy":1100},"sales_tiers":[{"min":0,"max":10000,"rate":0.02,"label":"1万以下"},{"min":10000,"max":15000,"rate":0.05,"label":"1万含-1.5万"},{"min":15000,"max":20000,"rate":0.08,"label":"1.5万含-2万"},{"min":20000,"max":25000,"rate":0.11,"label":"2万含-2.5万"},{"min":25000,"max":30000,"rate":0.12,"label":"2.5万含-3万"},{"min":30000,"max":0,"rate":0.13,"label":"3万及以上"}],"verify_rate":0.05},"staff":{"fixed_items":{"base_salary":1200,"overtime":700,"non_compete":0,"social_subsidy":1100},"sales_tiers":[{"min":0,"max":10000,"rate":0.02,"label":"1万以下"},{"min":10000,"max":15000,"rate":0.05,"label":"1万含-1.5万"},{"min":15000,"max":20000,"rate":0.08,"label":"1.5万含-2万"},{"min":20000,"max":25000,"rate":0.11,"label":"2万含-2.5万"},{"min":25000,"max":30000,"rate":0.12,"label":"2.5万含-3万"},{"min":30000,"max":0,"rate":0.13,"label":"3万及以上"}],"verify_rate":0.05},"manager_extra":{"min_store_share":0.10,"min_shortfall_penalty":0.08,"attend_days_gt":20,"cap_by_headcount":[{"max_people":2,"cap_ratio":0.6},{"people":3,"cap_ratio":0.4},{"min_people":4,"cap_ratio":0.3}]}}'
WHERE id=1;

-- +goose Down
ALTER TABLE lp_salary_rule DROP COLUMN config_json;
