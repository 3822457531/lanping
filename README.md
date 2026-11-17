# 兰瓶美业连锁管理系统

Go (Gin) + MySQL 8 + Vue 3 商业化重建版。旧静态原型见 [`lp/`](lp/)，方案文档生成器见 [`lanping/`](lanping/)。

## 架构

| 端 | 目录 | 技术 | 端口 |
|----|------|------|------|
| API | `server/` | Go Gin + sqlx + goose | 8080 |
| 管理后台 PC | `apps/admin/` | Vue3 + Element Plus | 5173 |
| 门店端 H5 | `apps/store/` | Vue3 + Vant | 5174 |
| 顾客 H5 | `apps/customer/` | Vue3 + Vant | 5175 |
| MySQL | docker-compose | MySQL 8 | 3306 |

## 快速启动

### 1. 启动 MySQL

**方式 A · Docker**

```bash
docker compose up -d mysql
```

**方式 B · 本机 MySQL 8**

```sql
CREATE DATABASE lanping DEFAULT CHARACTER SET utf8mb4;
CREATE USER 'lanping'@'%' IDENTIFIED BY 'lanping';
GRANT ALL ON lanping.* TO 'lanping'@'%';
FLUSH PRIVILEGES;
```

Windows 本地已初始化实例可用：

```powershell
.\scripts\start-mysql.ps1
```

默认已对接本机 phpstudy MySQL（`root/123456`，库名 `lanping`），见 [`server/configs/config.yaml`](server/configs/config.yaml)。

也可用：`lanping:lanping@tcp(127.0.0.1:3306)/lanping?parseTime=true&loc=Local&charset=utf8mb4`

### 2. 启动 API

```bash
cd server
go run ./cmd/api
# 或
go build -o bin/lp-api.exe ./cmd/api && .\bin\lp-api.exe
```

启动时若 `migrate.auto=true`（默认），会自动执行 `migrations/`。

健康检查：`GET http://127.0.0.1:8080/health`

### 3. 启动前端

```bash
cd apps/admin && npm install && npm run dev
cd apps/store && npm install && npm run dev
cd apps/customer && npm install && npm run dev
```

## 演示账号

| 角色 | 手机号 | 密码 |
|------|--------|------|
| 管理员 | 13800000001 | 123456 |
| 店长 | 13800000002 | 123456 |
| 员工 | 13800000004 | 123456 |
| 会员 | 13900000001 | （顾客端仅手机号登录） |

## 环境变量

| 变量 | 说明 |
|------|------|
| `LP_DB_DSN` | MySQL DSN |
| `LP_JWT_SECRET` | JWT 密钥 |
| `LP_HTTP_ADDR` | 监听地址，默认 `:8080` |
| `LP_AUTO_MIGRATE` | `true/false` 是否自动迁移 |
| `LP_CONFIG` | 配置文件路径 |

## 文档

- [API 说明](docs/api.md)
- [部署说明](docs/deploy.md)
