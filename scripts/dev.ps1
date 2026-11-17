# 本地开发一键提示（需已启动 MySQL）
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot

Write-Host "==> API :8080"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$root\server'; go run ./cmd/api"

Start-Sleep -Seconds 1
Write-Host "==> Admin :5173"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$root\apps\admin'; npm run dev"

Write-Host "==> Store :5174"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$root\apps\store'; npm run dev"

Write-Host "==> Customer :5175"
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$root\apps\customer'; npm run dev"

Write-Host "演示账号: 管理员 13800000001 / 店长 13800000002 / 员工 13800000004 ，密码 123456"
