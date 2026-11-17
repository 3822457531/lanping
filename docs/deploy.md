# 部署说明

## 推荐拓扑

```text
Nginx
  /api/*     → 127.0.0.1:8080
  /admin/*   → apps/admin/dist
  /store/*   → apps/store/dist
  /c/*       → apps/customer/dist
MySQL 8（独立主机或同机）
```

## 构建

```bash
# API
cd server && CGO_ENABLED=0 go build -o lp-api ./cmd/api

# 前端
cd apps/admin && npm ci && npm run build
cd apps/store && npm ci && npm run build
cd apps/customer && npm ci && npm run build
```

生产环境务必设置：

```bash
export LP_JWT_SECRET='强随机密钥'
export LP_DB_DSN='user:pass@tcp(db:3306)/lanping?parseTime=true&loc=Local&charset=utf8mb4'
export LP_AUTO_MIGRATE=false   # 生产建议显式执行 goose up
export LP_HTTP_ADDR=':8080'
```

## 迁移与备份

```bash
# 使用 goose CLI 或启动时 auto migrate
# 备份
mysqldump -u lanping -p lanping > backup-$(date +%F).sql
```

## Nginx 片段

```nginx
location /api/ {
  proxy_pass http://127.0.0.1:8080/api/;
  proxy_set_header Authorization $http_authorization;
  proxy_set_header X-Real-IP $remote_addr;
}
location / {
  root /var/www/lp-admin;
  try_files $uri $uri/ /index.html;
}
```

短信与 OSS：当前为 mock（写 `lp_sms_log`），上线时替换 `service` 内发送实现即可。
