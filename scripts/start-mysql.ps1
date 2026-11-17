# Start local MySQL for Lanping (initialized once under D:\mysql-lanping-data)
$ErrorActionPreference = "Stop"
$mysqld = "C:\Program Files\MySQL\MySQL Server 8.4\bin\mysqld.exe"
$conf = "D:\mysql-lanping-conf\my.ini"

if (-not (Test-Path $mysqld)) {
  Write-Error "MySQL not found at $mysqld. Install MySQL 8 or use: docker compose up -d mysql"
}
if (-not (Test-Path $conf)) {
  Write-Error "Missing $conf. See README for first-time init."
}

if (-not (Get-Process mysqld -ErrorAction SilentlyContinue)) {
  Start-Process -FilePath $mysqld -ArgumentList "--defaults-file=$conf" -WindowStyle Hidden
  Write-Host "mysqld starting..."
  Start-Sleep -Seconds 4
} else {
  Write-Host "mysqld already running"
}

$mysql = "C:\Program Files\MySQL\MySQL Server 8.4\bin\mysql.exe"
& $mysql --protocol=TCP --host=127.0.0.1 --port=3306 -ulanping -planping -e "SELECT 1 AS ok;" lanping | Out-Host
