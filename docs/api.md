# API 约定

统一响应：

```json
{ "ok": true, "data": {} }
{ "ok": false, "error": "原因" }
```

鉴权：`Authorization: Bearer <jwt>`

## 主要路由

### 认证
- `POST /api/auth/login` `{phone,password}` → `{token,user}`
- `POST /api/auth/register` 提交入职审核
- `GET /api/auth/me`
- `POST /api/customer/login` `{phone}`

### 经营
- `GET /api/dashboard` / `GET /api/workbench`
- `GET|POST|PUT|DELETE /api/customers`
- `GET|POST|PUT|DELETE /api/projects`
- `GET|POST /api/cards`
- `GET /api/verifications`
- `POST /api/verifications/draft`
- `POST /api/verifications/:id/confirm|cancel|refund`
- `GET /api/public/verification/:token`
- `POST /api/public/verification/:token/confirm`

### 运营
- `GET|POST|PUT|DELETE /api/appointments`
- `GET /api/inventory/stock` · `POST /api/inventory/orders`
- `POST /api/attendance/check` · `GET /api/attendance/today`
- `GET /api/audits` · `POST /api/audits/:id/approve|reject`
- `GET|PUT /api/settings`
- `GET /api/assessments`
- `GET /api/salary/rules` · `POST /api/salary/compute` · `GET /api/salary/records`
- `GET /api/welfare` · `POST /api/welfare/:id/issue`
- `GET /api/messages` · `GET /api/sms/logs` · `GET /api/operate-logs`

分页参数：`page`、`pageSize`（默认 20，上限 200）。

角色隔离：`manager/staff` 强制本店 `store_id`；`admin` 可跨店。
