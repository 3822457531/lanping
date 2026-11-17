# 兰瓶宝塔部署包

生成时间见目录内文件。含：Linux API、管理端、员工端、完整 SQL。不含顾客端。

## 目录

- api/lp-api          Linux amd64 后端（无后缀可执行文件）
- api/configs/        配置（上线请改 DSN / JWT）
- api/migrations/     迁移文件（首次可用 SQL 导入，不必再跑）
- admin/              管理端静态资源
- store/              员工/店长端静态资源
- sql/lanping_full.sql 完整库（含库名、表结构、演示数据）

## 1. 导入数据库

宝塔 → 数据库 → 导入 `sql/lanping_full.sql`
或命令行：

```bash
mysql -uroot -p < sql/lanping_full.sql
```

SQL 已包含 `CREATE DATABASE lanping` 与 `USE lanping`。

## 2. 启动 API

上传整个 `api/` 到服务器，例如 `/www/wwwroot/lanping/api/`：

```bash
chmod +x lp-api
# 建议用环境变量覆盖配置
export LP_DB_DSN='root:你的密码@tcp(127.0.0.1:3306)/lanping?parseTime=true&loc=Local&charset=utf8mb4'
export LP_JWT_SECRET='请改成强随机密钥'
export LP_HTTP_ADDR=':8080'
export LP_AUTO_MIGRATE='false'
./lp-api
```

宝塔可用「Supervisor / 进程守护」守护 `./lp-api`。

健康检查：`http://服务器IP:8080/api/health`

## 3. 前端站点（两个）

| 站点 | 根目录 | 说明 |
|------|--------|------|
| admin.xxx.com | deploy/admin | 管理端 |
| store.xxx.com | deploy/store | 员工/店长端 |

每个站点 Nginx 增加：

```nginx
location / {
  try_files $uri $uri/ /index.html;
}
location /api/ {
  proxy_pass http://127.0.0.1:8080/api/;
  proxy_set_header Host $host;
  proxy_set_header X-Real-IP $remote_addr;
  proxy_set_header Authorization $http_authorization;
}
```

## 演示账号

| 角色 | 手机号 | 密码 |
|------|--------|------|
| 管理员 | 13800000001 | 123456 |
| 店长 | 13800000002 | 123456 |
| 员工 | 13800000004 | 123456 |

上线后请立刻修改密码与 JWT 密钥。
