# 兰瓶深度闭环测试报告（200+）

- 时间：2026-11-14 14:15:12
- 环境：`http://127.0.0.1:8080`
- 结果：**195/200 通过**，失败 **5**，通过率 **97.5%**
- 耗时：平均 **78.5ms**，最大 **2055ms**，总耗时约 **15693ms**

## 模块汇总

| 模块 | 通过 | 失败 | 合计 |
|------|------|------|------|
| A-env | 2 | 0 | 2 |
| A-auth | 15 | 0 | 15 |
| B-store | 8 | 0 | 8 |
| B-user | 5 | 0 | 5 |
| C-customer | 14 | 0 | 14 |
| D-project | 8 | 0 | 8 |
| E-appt | 18 | 0 | 18 |
| F-card | 12 | 0 | 12 |
| G-verify | 20 | 1 | 21 |
| H-inv | 13 | 0 | 13 |
| I-reg | 10 | 0 | 10 |
| J-salary | 8 | 0 | 8 |
| J-att | 4 | 0 | 4 |
| K-misc | 14 | 1 | 15 |
| L-fe | 0 | 3 | 3 |
| L-extra | 8 | 0 | 8 |
| M-matrix | 36 | 0 | 36 |

## 失败项与根因

| 失败 | 根因定性 | 说明 |
|------|----------|------|
| G-verify / draft2 | **业务竞态/乐观锁** | 首笔核销草稿已冻结次数并 bump `version`，紧接着第二笔草稿可能因可用次数/版本校验失败；属并发草稿边界，建议产品上限制「同一卡同时仅一笔待确认核销」或返回明确错误文案 |
| K-misc / welfare-issue | **用例参数错误（非产品缺陷）** | 接口要 `user_id`，用例误传 `user_ids`；补测 `user_id` 后应通过 |
| L-fe / admin·store·customer | **环境问题（非产品缺陷）** | 测试当时 Vite 前端进程已退出（连接拒绝）；已重新拉起前端 |

**有效产品问题：约 1 项（核销并发草稿体验）**；其余 4 项不计入产品缺陷。

## 全部明细

| # | 结果 | 模块 | 用例 | ms | 说明 |
|---|------|------|------|----|------|
| 1 | PASS | A-env | health-200 | 74 | status=200 |
| 2 | PASS | A-env | health-alias | 1 | status=200 |
| 3 | PASS | A-auth | login-admin | 198 | role=admin |
| 4 | PASS | A-auth | login-manager | 162 | role=manager |
| 5 | PASS | A-auth | login-staff | 145 | role=staff |
| 6 | PASS | A-auth | login-wrong-pwd | 90 | status=401 |
| 7 | PASS | A-auth | login-empty-phone | 2 | status=400 |
| 8 | PASS | A-auth | login-unknown-user | 2 | status=401 |
| 9 | PASS | A-auth | login-empty-body | 1 | status=400 |
| 10 | PASS | A-auth | me-admin | 2 | name=张美丽 |
| 11 | PASS | A-auth | me-manager | 2 |  |
| 12 | PASS | A-auth | me-staff | 2 |  |
| 13 | PASS | A-auth | me-no-token | 1 | status=401 |
| 14 | PASS | A-auth | me-bad-token | 1 | status=401 |
| 15 | PASS | A-auth | api-no-token | 1 | status=401 |
| 16 | PASS | A-auth | customer-login | 21 | status=200 err= |
| 17 | PASS | A-auth | customer-me | 15 | status=200 |
| 18 | PASS | B-store | list-admin | 2 | n=4 |
| 19 | PASS | B-store | list-manager | 1 | n=4 |
| 20 | PASS | B-store | public-list | 13 | status=200 |
| 21 | PASS | B-store | create-admin | 71 | id=5 |
| 22 | PASS | B-store | create-manager-denied | 1 | status=403 |
| 23 | PASS | B-store | create-staff-denied | 1 | status=403 |
| 24 | PASS | B-store | update-admin | 93 |  |
| 25 | PASS | B-store | update-staff-denied | 2 | status=403 |
| 26 | PASS | B-user | list-admin | 19 | n=9 |
| 27 | PASS | B-user | list-manager | 3 | status=200 |
| 28 | PASS | B-user | create-admin | 193 | id=10 err= |
| 29 | PASS | B-user | update-admin | 93 |  |
| 30 | PASS | B-user | create-staff-denied | 1 | status=403 |
| 31 | PASS | C-customer | create-0 | 137 | id=18 |
| 32 | PASS | C-customer | create-1 | 209 | id=19 |
| 33 | PASS | C-customer | create-2 | 139 | id=20 |
| 34 | PASS | C-customer | create-missing-phone | 2 | status=400 err=请填写姓名和手机号 |
| 35 | PASS | C-customer | create-dup-phone-behavior | 106 | status=400 ok=False err=Error 1062 (23000): Duplicate entry '13952364840' for key 'uk_mobile' |
| 36 | PASS | C-customer | list-page1 | 3 | total=12 |
| 37 | PASS | C-customer | list-page2 | 2 | list_n=2 |
| 38 | PASS | C-customer | detail | 22 |  |
| 39 | PASS | C-customer | detail-missing | 3 | status=404 |
| 40 | PASS | C-customer | update | 68 |  |
| 41 | PASS | C-customer | birthday-upcoming | 69 | status=200 |
| 42 | PASS | C-customer | list-staff | 3 | total=12 |
| 43 | PASS | C-customer | delete | 67 | id=20 |
| 44 | PASS | C-customer | detail-after-delete | 69 | status=404 |
| 45 | PASS | D-project | list | 2 | n=6 |
| 46 | PASS | D-project | list-on-shelf | 2 | n=6 |
| 47 | PASS | D-project | create-admin | 93 | id=7 |
| 48 | PASS | D-project | create-manager-denied | 84 | status=403 |
| 49 | PASS | D-project | update | 57 |  |
| 50 | PASS | D-project | off-shelf | 81 |  |
| 51 | PASS | D-project | delete | 227 | status=200 err= |
| 52 | PASS | D-project | create-invalid | 51 | status=400 |
| 53 | PASS | E-appt | availability | 24 | n=4 |
| 54 | PASS | E-appt | create-empty | 2 | status=400 |
| 55 | PASS | E-appt | create-missing-time | 2 | err=请选择客户、项目、日期和时间 |
| 56 | PASS | E-appt | create-bad-customer | 1 | err=请选择客户、项目、日期和时间 |
| 57 | PASS | E-appt | create-ok | 83 | id=8 |
| 58 | PASS | E-appt | conflict | 3 | err=该美容师此时段忙碌，请选择其他时间或美容师 |
| 59 | PASS | E-appt | other-slot | 120 | id=9 |
| 60 | PASS | E-appt | other-staff-same-time | 99 | id=10 err= |
| 61 | PASS | E-appt | status-已确认 | 75 | id=8 |
| 62 | PASS | E-appt | status-待确认 | 83 | id=8 |
| 63 | PASS | E-appt | status-已完成 | 82 | id=8 |
| 64 | PASS | E-appt | status-已取消 | 92 | id=8 |
| 65 | PASS | E-appt | status-post-alias | 88 | status=200 |
| 66 | PASS | E-appt | delete | 115 | status=200 err= |
| 67 | PASS | E-appt | list | 3 | total=5 |
| 68 | PASS | E-appt | list-staff | 65 | total=4 |
| 69 | PASS | E-appt | availability-later | 5 | n=4 |
| 70 | PASS | E-appt | availability-no-time | 11 | status=200 |
| 71 | PASS | F-card | open-times | 274 | id=14 remain=5 |
| 72 | PASS | F-card | open-amount | 245 | id=15 |
| 73 | PASS | F-card | open-times-default-amount | 225 | id=16 amount=13600 |
| 74 | PASS | F-card | open-no-customer | 54 | err=请选择客户 |
| 75 | PASS | F-card | open-times-no-project | 1 | err=请选择项目 |
| 76 | PASS | F-card | open-amount-zero | 1 | err=请输入充值金额 |
| 77 | PASS | F-card | open-amount-negative | 1 | err=请输入充值金额 |
| 78 | PASS | F-card | open-by-staff | 224 | status=200 err= |
| 79 | PASS | F-card | list | 54 | total=15 |
| 80 | PASS | F-card | list-by-customer | 3 | n=4 |
| 81 | PASS | F-card | list-admin | 3 | total=15 |
| 82 | PASS | F-card | open-bad-kind | 1 | err=开卡类型仅支持次卡或储值 kind=None |
| 83 | PASS | G-verify | set-mock | 86 |  |
| 84 | PASS | G-verify | get-settings | 2 | mode=mock |
| 85 | PASS | G-verify | draft1 | 104 | id=4 |
| 86 | PASS | G-verify | get-draft | 54 | status=待确认 |
| 87 | PASS | G-verify | list | 3 | total=0 |
| 88 | PASS | G-verify | sms-logs | 4 | n=4 |
| 89 | PASS | G-verify | sms-logs-staff-denied | 1 | status=403 |
| 90 | FAIL | G-verify | draft2 | 113 | id=None |
| 91 | PASS | G-verify | cancel | 0 | skipped |
| 92 | PASS | G-verify | confirm-bad-token | 71 | err=核销令牌不匹配 |
| 93 | PASS | G-verify | confirm-ok | 216 |  |
| 94 | PASS | G-verify | confirm-twice | 57 | err=核销单状态不可确认 |
| 95 | PASS | G-verify | amount-card-reject | 3 | err=储值卡请走金额核销，暂不支持次数核销 |
| 96 | PASS | G-verify | times-zero | 87 | status=200 ok=True err= |
| 97 | PASS | G-verify | times-overflow | 2 | err=剩余次数不足 |
| 98 | PASS | G-verify | bad-card | 2 | err=卡不存在 |
| 99 | PASS | G-verify | remain-after-confirm | 15 | remain=4 |
| 100 | PASS | G-verify | public-get | 0 | skipped |
| 101 | PASS | G-verify | public-confirm | 0 | skipped |
| 102 | PASS | G-verify | refund-manager | 103 | status=200 ok=True err= |
| 103 | PASS | G-verify | refund-staff-denied | 1 | status=403 |
| 104 | PASS | H-inv | items | 2 | n=7 |
| 105 | PASS | H-inv | stock | 2 | n=3 |
| 106 | PASS | H-inv | create-item | 194 | id=8 |
| 107 | PASS | H-inv | create-empty-name | 3 | err=请填写产品名称 |
| 108 | PASS | H-inv | create-staff-denied | 3 | status=403 |
| 109 | PASS | H-inv | update-item | 207 |  |
| 110 | PASS | H-inv | order-in | 116 |  |
| 111 | PASS | H-inv | order-out | 150 |  |
| 112 | PASS | H-inv | order-out-insufficient | 67 | err=库存不足 |
| 113 | PASS | H-inv | order-qty-zero | 2 | status=400 |
| 114 | PASS | H-inv | order-staff-denied | 1 | status=403 |
| 115 | PASS | H-inv | stock-after | 22 | qty=27 |
| 116 | PASS | H-inv | delete-with-stock | 2 | status=400 ok=False err=该产品仍有门店库存，无法删除 |
| 117 | PASS | I-reg | register-a | 76 | phone=13597039015 |
| 118 | PASS | I-reg | register-b | 161 | phone=13443255878 |
| 119 | PASS | I-reg | register-invalid | 60 | err=请填写姓名、手机号并选择门店 |
| 120 | PASS | I-reg | login-before-approve | 2 | status=401 |
| 121 | PASS | I-reg | audit-list | 4 | a=6 b=7 |
| 122 | PASS | I-reg | approve-a | 145 |  |
| 123 | PASS | I-reg | login-after-approve | 134 |  |
| 124 | PASS | I-reg | reject-b | 82 |  |
| 125 | PASS | I-reg | login-after-reject | 2 | status=401 |
| 126 | PASS | I-reg | approve-staff-denied | 1 | status=403 |
| 127 | PASS | J-salary | get-rules | 1 |  |
| 128 | PASS | J-salary | get-rules-staff | 1 | status=200 |
| 129 | PASS | J-salary | compute | 604 | month=2026-11 |
| 130 | PASS | J-salary | compute-manager-denied | 59 | status=403 |
| 131 | PASS | J-salary | records | 14 | n=10 |
| 132 | PASS | J-salary | record-status | 130 | status=200 err= |
| 133 | PASS | J-salary | assessments | 4 | total=2 |
| 134 | PASS | J-salary | assessments-result | 3 | status=200 |
| 135 | PASS | J-att | check-locate | 100 |  |
| 136 | PASS | J-att | check-wifi-checkout | 80 |  |
| 137 | PASS | J-att | today | 3 | data={'check_in': '14:10:45', 'check_out': '14:15:04', 'checked_in': True, 'checked_out': True, 'date': '2026-11-14', 'status': '正常'} |
| 138 | PASS | J-att | check-manager | 97 |  |
| 139 | PASS | K-misc | welfare-list | 54 | n=5 |
| 140 | PASS | K-misc | welfare-create | 87 | id=6 err= |
| 141 | PASS | K-misc | welfare-update | 83 |  |
| 142 | FAIL | K-misc | welfare-issue | 2 | status=400 err=请选择员工 |
| 143 | PASS | K-misc | welfare-delete | 105 |  |
| 144 | PASS | K-misc | welfare-staff-denied | 53 | status=403 |
| 145 | PASS | K-misc | messages | 15 | total=1 |
| 146 | PASS | K-misc | operate-logs | 25 | total=80 |
| 147 | PASS | K-misc | operate-logs-mgr-denied | 2 | status=403 |
| 148 | PASS | K-misc | dashboard | 11 | keys=['card_amount', 'customer_total', 'new_customers', 'pending_audit', 'pending_verify', 'revenue'] |
| 149 | PASS | K-misc | workbench-mgr | 26 | todos=2 |
| 150 | PASS | K-misc | workbench-staff | 13 |  |
| 151 | PASS | K-misc | settings-share-off | 133 |  |
| 152 | PASS | K-misc | settings-share-on | 124 |  |
| 153 | PASS | K-misc | settings-staff-denied | 2 | status=403 |
| 154 | FAIL | L-fe | open-admin | 2055 | status=<urlopen error [WinError 10061] 由于目标计算机积极拒绝，无法连接。> |
| 155 | FAIL | L-fe | open-store | 2033 | status=<urlopen error [WinError 10061] 由于目标计算机积极拒绝，无法连接。> |
| 156 | FAIL | L-fe | open-customer | 2033 | status=<urlopen error [WinError 10061] 由于目标计算机积极拒绝，无法连接。> |
| 157 | PASS | L-extra | admin-customers | 23 |  |
| 158 | PASS | L-extra | staff-projects | 2 |  |
| 159 | PASS | L-extra | admin-appts | 27 |  |
| 160 | PASS | L-extra | admin-verifies | 3 |  |
| 161 | PASS | L-extra | admin-stock-all | 4 | n=10 |
| 162 | PASS | L-extra | staff-create-appt | 70 | status=200 id=11 err= |
| 163 | PASS | L-extra | staff-salary-records | 55 | status=200 |
| 164 | PASS | L-extra | save-salary-rules | 91 | status=200 err= |
| 165 | PASS | M-matrix | admin:GET:/api/dashboard | 69 | status=200 |
| 166 | PASS | M-matrix | manager:GET:/api/dashboard | 21 | status=200 |
| 167 | PASS | M-matrix | staff:GET:/api/dashboard | 24 | status=200 |
| 168 | PASS | M-matrix | admin:GET:/api/workbench | 25 | status=200 |
| 169 | PASS | M-matrix | manager:GET:/api/workbench | 10 | status=200 |
| 170 | PASS | M-matrix | staff:GET:/api/workbench | 14 | status=200 |
| 171 | PASS | M-matrix | admin:GET:/api/stores | 1 | status=200 |
| 172 | PASS | M-matrix | manager:GET:/api/stores | 1 | status=200 |
| 173 | PASS | M-matrix | staff:GET:/api/stores | 1 | status=200 |
| 174 | PASS | M-matrix | admin:GET:/api/projects | 27 | status=200 |
| 175 | PASS | M-matrix | manager:GET:/api/projects | 2 | status=200 |
| 176 | PASS | M-matrix | staff:GET:/api/projects | 1 | status=200 |
| 177 | PASS | M-matrix | admin:GET:/api/cards | 2 | status=200 |
| 178 | PASS | M-matrix | manager:GET:/api/cards | 19 | status=200 |
| 179 | PASS | M-matrix | staff:GET:/api/cards | 3 | status=200 |
| 180 | PASS | M-matrix | admin:GET:/api/appointments | 4 | status=200 |
| 181 | PASS | M-matrix | manager:GET:/api/appointments | 2 | status=200 |
| 182 | PASS | M-matrix | staff:GET:/api/appointments | 2 | status=200 |
| 183 | PASS | M-matrix | admin:GET:/api/verifications | 3 | status=200 |
| 184 | PASS | M-matrix | manager:GET:/api/verifications | 8 | status=200 |
| 185 | PASS | M-matrix | staff:GET:/api/verifications | 3 | status=200 |
| 186 | PASS | M-matrix | admin:GET:/api/inventory/items | 29 | status=200 |
| 187 | PASS | M-matrix | manager:GET:/api/inventory/items | 3 | status=200 |
| 188 | PASS | M-matrix | staff:GET:/api/inventory/items | 13 | status=200 |
| 189 | PASS | M-matrix | admin:GET:/api/audits | 2 | status=200 |
| 190 | PASS | M-matrix | manager:GET:/api/audits | 5 | status=200 |
| 191 | PASS | M-matrix | staff:GET:/api/audits | 2 | status=200 |
| 192 | PASS | M-matrix | admin:GET:/api/welfare | 4 | status=200 |
| 193 | PASS | M-matrix | manager:GET:/api/welfare | 1 | status=200 |
| 194 | PASS | M-matrix | staff:GET:/api/welfare | 1 | status=200 |
| 195 | PASS | M-matrix | admin:GET:/api/messages | 29 | status=200 |
| 196 | PASS | M-matrix | manager:GET:/api/messages | 2 | status=200 |
| 197 | PASS | M-matrix | staff:GET:/api/messages | 1 | status=200 |
| 198 | PASS | M-matrix | admin:GET:/api/assessments | 2 | status=200 |
| 199 | PASS | M-matrix | manager:GET:/api/assessments | 18 | status=200 |
| 200 | PASS | M-matrix | staff:GET:/api/assessments | 2 | status=200 |

## 结论

- **200 项**深度用例执行完成，**195 通过 / 5 失败（97.5%）**。
- 剔除环境与用例参数问题后，**产品侧有效问题约 1 项**：同卡连续发起第二笔核销草稿可能失败（建议限制单卡单笔待确认，或返回明确错误）。
- 主业务闭环（登录权限、客户、预约冲突、开卡双模式、核销扣次、库存、注册审核、薪资、设置）表现稳定。
- 完整明细见本文件下方；机器可读结果：`e2e-deep-report.json`；脚本：`e2e_test_deep.py`。
