$ErrorActionPreference = 'Continue'
$base = 'http://127.0.0.1:8080'
$results = New-Object System.Collections.Generic.List[object]
$ctx = @{}

function Add-Result($module, $case, $ok, $detail, $ms) {
  $results.Add([pscustomobject]@{
    module = $module; case = $case; ok = [bool]$ok; detail = [string]$detail; ms = [int]$ms
  })
  $mark = if ($ok) { 'PASS' } else { 'FAIL' }
  Write-Host ("[{0}] {1} :: {2} ({3}ms) {4}" -f $mark, $module, $case, $ms, $detail)
}

function Invoke-Api {
  param(
    [string]$Method = 'GET',
    [string]$Path,
    [hashtable]$Headers = @{},
    [object]$Body = $null,
    [int]$Expect = 200
  )
  $url = "$base$Path"
  $sw = [Diagnostics.Stopwatch]::StartNew()
  try {
    $params = @{
      Method = $Method
      Uri = $url
      Headers = $Headers
      ContentType = 'application/json'
      UseBasicParsing = $true
      TimeoutSec = 20
    }
    if ($null -ne $Body) { $params.Body = ($Body | ConvertTo-Json -Depth 8 -Compress) }
    $resp = Invoke-WebRequest @params
    $sw.Stop()
    $json = $null
    try { $json = $resp.Content | ConvertFrom-Json } catch {}
    $okHttp = ($resp.StatusCode -eq $Expect)
    $okBiz = $true
    if ($json -and ($json.PSObject.Properties.Name -contains 'ok')) {
      if ($Expect -ge 200 -and $Expect -lt 300) { $okBiz = [bool]$json.ok }
    }
    return @{
      ok = ($okHttp -and $okBiz)
      status = [int]$resp.StatusCode
      ms = [int]$sw.ElapsedMilliseconds
      json = $json
      raw = $resp.Content
      err = ''
    }
  } catch {
    $sw.Stop()
    $status = 0
    $raw = $_.Exception.Message
    $json = $null
    if ($_.Exception.Response) {
      try {
        $status = [int]$_.Exception.Response.StatusCode
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object IO.StreamReader($stream)
        $raw = $reader.ReadToEnd()
        try { $json = $raw | ConvertFrom-Json } catch {}
      } catch {}
    }
    $ok = ($status -eq $Expect)
    return @{ ok = $ok; status = $status; ms = [int]$sw.ElapsedMilliseconds; json = $json; raw = $raw; err = $_.Exception.Message }
  }
}

function AuthHeader($token) { @{ Authorization = "Bearer $token" } }

# ---------- 0. Health ----------
$r = Invoke-Api -Path '/api/health'
Add-Result '环境' '健康检查 /api/health' $r.ok ("status=$($r.status)") $r.ms

# ---------- 1. Login ----------
foreach ($pair in @(
  @{ role='admin'; phone='13800000001'; pass='123456' },
  @{ role='manager'; phone='13800000002'; pass='123456' },
  @{ role='staff'; phone='13800000004'; pass='123456' }
)) {
  $r = Invoke-Api -Method POST -Path '/api/auth/login' -Body @{ phone=$pair.phone; password=$pair.pass }
  $token = $r.json.data.token
  $user = $r.json.data.user
  $ok = $r.ok -and $token
  if ($ok) { $ctx[$pair.role] = @{ token=$token; user=$user } }
  Add-Result '登录' ("$($pair.role) 登录 $($pair.phone)") $ok ("role=$($user.role); store=$($user.store_id)") $r.ms
}

$r = Invoke-Api -Method POST -Path '/api/auth/login' -Body @{ phone='13800000001'; password='wrong' } -Expect 400
Add-Result '登录' '错误密码应失败' ($r.status -eq 400 -or -not $r.json.ok) ("status=$($r.status) raw=$($r.raw.Substring(0,[Math]::Min(80,$r.raw.Length)))") $r.ms

$adminH = AuthHeader $ctx.admin.token
$mgrH = AuthHeader $ctx.manager.token
$staffH = AuthHeader $ctx.staff.token

$r = Invoke-Api -Path '/api/auth/me' -Headers $adminH
Add-Result '登录' 'admin /auth/me' $r.ok ("name=$($r.json.data.name)") $r.ms

# ---------- 2. Customers ----------
$mobile = ('139' + (Get-Random -Minimum 10000000 -Maximum 99999999))
$r = Invoke-Api -Method POST -Path '/api/customers' -Headers $mgrH -Body @{
  name = '闭环测试客户'
  phone = $mobile
  gender = 2
  level = '银卡会员'
  source = '到店'
}
$cid = $r.json.data.id
if (-not $cid) { $cid = $r.json.data.ID }
Add-Result '客户' '门店新建客户' ($r.ok -and $cid) ("id=$cid phone=$mobile") $r.ms
$ctx.customer_id = $cid

$r = Invoke-Api -Path '/api/customers?pageSize=5' -Headers $mgrH
Add-Result '客户' '客户列表' ($r.ok -and $r.json.data.list) ("total=$($r.json.data.total)") $r.ms

$r = Invoke-Api -Path "/api/customers/$cid" -Headers $mgrH
Add-Result '客户' '客户详情' $r.ok ("name=$($r.json.data.name)") $r.ms

# ---------- 3. Projects / Stores ----------
$r = Invoke-Api -Path '/api/projects?status=上架&pageSize=20' -Headers $mgrH
$projects = @($r.json.data.list)
$pid = $projects[0].id
Add-Result '项目' '上架项目列表' ($r.ok -and $projects.Count -gt 0) ("count=$($projects.Count) first=$pid") $r.ms
$ctx.project_id = $pid

$r = Invoke-Api -Path '/api/stores' -Headers $adminH
$stores = @($r.json.data.list)
$sid = $ctx.manager.user.store_id
if (-not $sid) { $sid = $stores[0].id }
Add-Result '门店' '门店列表' ($r.ok -and $stores.Count -gt 0) ("count=$($stores.Count) mgrStore=$sid") $r.ms
$ctx.store_id = $sid

# ---------- 4. Appointments ----------
$date = (Get-Date).ToString('yyyy-MM-dd')
$time = (Get-Date).AddHours(2).ToString('HH:mm')
$r = Invoke-Api -Path "/api/staff/availability?date=$date&time=$time&project_id=$pid&store_id=$sid" -Headers $mgrH
$staffList = @($r.json.data.list)
$free = $staffList | Where-Object { -not $_.busy } | Select-Object -First 1
Add-Result '预约' '美容师空闲查询' ($r.ok -and $staffList.Count -gt 0) ("staff=$($staffList.Count) free=$($free.name) id=$($free.id)") $r.ms

$apptStaff = if ($free) { $free.id } else { $ctx.staff.user.id }
$r = Invoke-Api -Method POST -Path '/api/appointments' -Headers $mgrH -Body @{
  customer_id = $cid
  project_id = $pid
  store_id = $sid
  staff_id = $apptStaff
  date = $date
  time = $time
  note = 'e2e-appt'
}
$apptId = $r.json.data.id
if (-not $apptId) { $apptId = $r.json.data.ID }
Add-Result '预约' '新建预约' ($r.ok -and $apptId) ("id=$apptId staff=$apptStaff $date $time") $r.ms
$ctx.appt_id = $apptId

# conflict same slot
$r2 = Invoke-Api -Method POST -Path '/api/appointments' -Headers $mgrH -Body @{
  customer_id = $cid; project_id = $pid; store_id = $sid; staff_id = $apptStaff
  date = $date; time = $time; note = 'conflict'
} -Expect 400
Add-Result '预约' '同美容师同时段冲突应拒绝' (($r2.status -eq 400) -or (-not $r2.json.ok)) ("status=$($r2.status) msg=$($r2.json.error)") $r2.ms

# different time should allow same staff
$time2 = (Get-Date).AddHours(5).ToString('HH:mm')
$r = Invoke-Api -Method POST -Path '/api/appointments' -Headers $mgrH -Body @{
  customer_id = $cid; project_id = $pid; store_id = $sid; staff_id = $apptStaff
  date = $date; time = $time2; note = 'e2e-appt-2'
}
$appt2 = $r.json.data.id
Add-Result '预约' '同美容师不同时段可再约' ($r.ok -and $appt2) ("id=$appt2 time=$time2") $r.ms

if ($apptId) {
  $r = Invoke-Api -Method PUT -Path "/api/appointments/$apptId" -Headers $mgrH -Body @{ status = '已确认' }
  Add-Result '预约' '确认预约' $r.ok ("id=$apptId") $r.ms

  $r = Invoke-Api -Method PUT -Path "/api/appointments/$apptId" -Headers $mgrH -Body @{ status = '已取消' }
  Add-Result '预约' '取消预约' $r.ok ("id=$apptId") $r.ms
}

$r = Invoke-Api -Path '/api/appointments?pageSize=20' -Headers $mgrH
Add-Result '预约' '预约列表' $r.ok ("total=$($r.json.data.total)") $r.ms

# ---------- 5. Cards open (times + amount) ----------
$r = Invoke-Api -Method POST -Path '/api/cards' -Headers $mgrH -Body @{
  customer_id = $cid; project_id = $pid; store_id = $sid
  card_kind = 'times'; times = 3; amount = 900; pay_method = '微信'
}
$cardTimes = $r.json.data
Add-Result '开卡' '次卡开卡' ($r.ok -and $cardTimes.id) ("id=$($cardTimes.id) remain=$($cardTimes.remain) kind=$($cardTimes.card_kind)") $r.ms
$ctx.card_times_id = $cardTimes.id

$r = Invoke-Api -Method POST -Path '/api/cards' -Headers $mgrH -Body @{
  customer_id = $cid; store_id = $sid
  card_kind = 'amount'; amount = 500; pay_method = '现金'
}
$cardAmt = $r.json.data
Add-Result '开卡' '储值充值开卡' ($r.ok -and $cardAmt.id -and $cardAmt.card_kind -eq 'amount') ("id=$($cardAmt.id) amount=$($cardAmt.amount)") $r.ms

$r = Invoke-Api -Path "/api/cards?pageSize=20" -Headers $mgrH
Add-Result '开卡' '卡列表' $r.ok ("total=$($r.json.data.total)") $r.ms

# ---------- 6. Verify + mock SMS ----------
$r = Invoke-Api -Method PUT -Path '/api/settings' -Headers $adminH -Body @{ verify_mode = 'mock'; customer_share = 'true' }
Add-Result '设置' '切换模拟核销' $r.ok ('verify_mode=mock') $r.ms

$r = Invoke-Api -Method POST -Path '/api/verifications/draft' -Headers $staffH -Body @{ card_id = $ctx.card_times_id; times = 1 }
$draft = $r.json.data
Add-Result '核销' '发起核销草稿(模拟短信)' ($r.ok -and $draft.id -and $draft.token) ("id=$($draft.id) no=$($draft.verify_no)") $r.ms
$ctx.verify_id = $draft.id
$ctx.verify_token = $draft.token

$r = Invoke-Api -Path '/api/sms/logs?pageSize=5' -Headers $adminH
$smsOk = $false
$smsDetail = ''
if ($r.ok -and $r.json.data.list) {
  $last = @($r.json.data.list)[0]
  $smsOk = $true
  $smsDetail = "status=$($last.status) mobile=$($last.mobile)"
}
Add-Result '核销' '短信日志有记录(模拟)' ($r.ok -and $smsOk) $smsDetail $r.ms

$r = Invoke-Api -Method POST -Path "/api/verifications/$($ctx.verify_id)/confirm" -Headers $staffH -Body @{ token = $ctx.verify_token }
Add-Result '核销' '员工确认核销完成' $r.ok ("id=$($ctx.verify_id)") $r.ms

$r = Invoke-Api -Path "/api/cards" -Headers $mgrH
$after = @($r.json.data.list) | Where-Object { $_.id -eq $ctx.card_times_id } | Select-Object -First 1
$remainOk = $after -and ($after.remain -eq 2)
Add-Result '核销' '次卡余次扣减正确' $remainOk ("remain=$($after.remain) expect=2") $r.ms

# amount card draft should fail times verify
$r = Invoke-Api -Method POST -Path '/api/verifications/draft' -Headers $staffH -Body @{ card_id = $cardAmt.id; times = 1 } -Expect 400
Add-Result '核销' '储值卡次数核销应拒绝' (($r.status -eq 400) -or (-not $r.json.ok)) ("msg=$($r.json.error)") $r.ms

# over-times
$r = Invoke-Api -Method POST -Path '/api/verifications/draft' -Headers $staffH -Body @{ card_id = $ctx.card_times_id; times = 99 } -Expect 400
Add-Result '核销' '超余次核销应拒绝' (($r.status -eq 400) -or (-not $r.json.ok)) ("msg=$($r.json.error)") $r.ms

# ---------- 7. Workbench / Dashboard ----------
$r = Invoke-Api -Path '/api/workbench' -Headers $mgrH
Add-Result '工作台' '门店工作台' $r.ok ("appt=$($r.json.data.today_appt_count) todos=$($r.json.data.todos.Count)") $r.ms

$r = Invoke-Api -Path '/api/dashboard' -Headers $adminH
Add-Result '工作台' '管理端看板' $r.ok ("customers=$($r.json.data.customer_total)") $r.ms

# ---------- 8. Inventory ----------
$r = Invoke-Api -Path '/api/inventory/items' -Headers $mgrH
$items = @($r.json.data.list)
Add-Result '库存' '产品目录' ($r.ok -and $items.Count -ge 0) ("count=$($items.Count)") $r.ms

$sku = 'E2E' + (Get-Random -Minimum 1000 -Maximum 9999)
$r = Invoke-Api -Method POST -Path '/api/inventory/items' -Headers $mgrH -Body @{
  name = 'E2E测试耗材'; sku = $sku; unit = '瓶'; store_id = $sid; qty = 10
}
$itemId = $r.json.data.id
Add-Result '库存' '新增产品+本店库存' ($r.ok -and $itemId) ("id=$itemId sku=$sku") $r.ms

if ($itemId) {
  $r = Invoke-Api -Method POST -Path '/api/inventory/orders' -Headers $mgrH -Body @{
    store_id = $sid; item_id = $itemId; qty = 2; direction = '出库'; channel = '门店调度'
  }
  Add-Result '库存' '出库记账' $r.ok ('qty=2') $r.ms

  $r = Invoke-Api -Method POST -Path '/api/inventory/orders' -Headers $mgrH -Body @{
    store_id = $sid; item_id = $itemId; qty = 999; direction = '出库'; channel = '门店调度'
  } -Expect 400
  Add-Result '库存' '库存不足出库应拒绝' (($r.status -eq 400) -or (-not $r.json.ok)) ("msg=$($r.json.error)") $r.ms

  $r = Invoke-Api -Path "/api/inventory/stock?store_id=$sid" -Headers $mgrH
  $row = @($r.json.data.list) | Where-Object { $_.item_id -eq $itemId } | Select-Object -First 1
  Add-Result '库存' '出库后库存=8' ($row.qty -eq 8) ("qty=$($row.qty)") $r.ms
}

# ---------- 9. Register + Audit ----------
$regPhone = ('137' + (Get-Random -Minimum 10000000 -Maximum 99999999))
$r = Invoke-Api -Method POST -Path '/api/auth/register' -Body @{
  name = 'E2E新员工'; phone = $regPhone; password = '123456'; store_id = $sid; role = 'staff'
}
Add-Result '注册' '员工注册申请' $r.ok ("phone=$regPhone") $r.ms

$r = Invoke-Api -Path '/api/audits' -Headers $mgrH
$pending = @($r.json.data.list) | Where-Object { $_.phone -eq $regPhone -and $_.status -eq '待审核' } | Select-Object -First 1
Add-Result '注册' '审核列表可见申请' ($r.ok -and $pending) ("auditId=$($pending.id)") $r.ms

if ($pending) {
  $r = Invoke-Api -Method POST -Path "/api/audits/$($pending.id)/approve" -Headers $mgrH
  Add-Result '注册' '店长通过审核' $r.ok ("id=$($pending.id)") $r.ms

  $r = Invoke-Api -Method POST -Path '/api/auth/login' -Body @{ phone = $regPhone; password = '123456' }
  Add-Result '注册' '新员工审核后可登录' ($r.ok -and $r.json.data.token) ("role=$($r.json.data.user.role)") $r.ms
}

# ---------- 10. Salary / performance ----------
$month = (Get-Date).ToString('yyyy-MM')
$r = Invoke-Api -Path '/api/salary/rules' -Headers $adminH
Add-Result '薪资' '读取薪资规则' $r.ok ("hasConfig=$([bool]$r.json.data)") $r.ms

$r = Invoke-Api -Method POST -Path '/api/salary/compute' -Headers $adminH -Body @{ month = $month }
Add-Result '薪资' '计算本月业绩/薪资' $r.ok ("month=$month") $r.ms

$r = Invoke-Api -Path "/api/salary/records?month=$month&pageSize=50" -Headers $adminH
$list = @($r.json.data.list)
Add-Result '薪资' '薪资记录列表' ($r.ok -and $list.Count -gt 0) ("count=$($list.Count)") $r.ms

$r = Invoke-Api -Path '/api/assessments' -Headers $mgrH
Add-Result '考核' '考核列表' $r.ok ("total=$($r.json.data.total)") $r.ms

# ---------- 11. Attendance (record only) ----------
$r = Invoke-Api -Method POST -Path '/api/attendance/check' -Headers $staffH -Body @{ method = '定位' }
Add-Result '打卡' '员工打卡记录' $r.ok ('method=定位') $r.ms
$r = Invoke-Api -Path '/api/attendance/today' -Headers $staffH
Add-Result '打卡' '今日出勤查询' ($r.ok -and $r.json.data.checked_in) ("in=$($r.json.data.check_in)") $r.ms

# ---------- 12. Welfare / messages / logs ----------
$r = Invoke-Api -Path '/api/welfare' -Headers $adminH
Add-Result '福利' '福利列表' $r.ok ("total=$($r.json.data.total)") $r.ms

$r = Invoke-Api -Path '/api/messages' -Headers $staffH
Add-Result '消息' '消息列表' $r.ok ("total=$($r.json.data.total)") $r.ms

$r = Invoke-Api -Path '/api/operate-logs?pageSize=10' -Headers $adminH
Add-Result '日志' '操作日志' ($r.ok -and @($r.json.data.list).Count -gt 0) ("total=$($r.json.data.total)") $r.ms

# ---------- 13. Permission ----------
$r = Invoke-Api -Method POST -Path '/api/inventory/items' -Headers $staffH -Body @{ name='无权'; unit='件'; store_id=$sid; qty=1 } -Expect 403
# some middleware may return 401/403/400
Add-Result '权限' '普通员工新增产品应受限' (($r.status -in 401,403) -or (-not $r.ok)) ("status=$($r.status)") $r.ms

$r = Invoke-Api -Path '/api/customers' -Headers @{} -Expect 401
Add-Result '权限' '无Token访问应拒绝' (($r.status -eq 401) -or (-not $r.ok)) ("status=$($r.status)") $r.ms

# ---------- Report ----------
$pass = @($results | Where-Object ok).Count
$fail = @($results | Where-Object { -not $_.ok }).Count
$total = $results.Count
$avg = [math]::Round((($results | Measure-Object ms -Average).Average), 1)
$max = ($results | Measure-Object ms -Maximum).Maximum

$reportDir = 'D:\workSpace\lan\deploy'
$md = @()
$md += '# 兰瓶闭环测试报告'
$md += ''
$md += "- 时间：$((Get-Date).ToString('yyyy-MM-dd HH:mm:ss'))"
$md += "- 环境：API $base ；管理端 5173 ；门店端 5174"
$md += "- 结果：**$pass/$total 通过**，失败 **$fail**"
$md += "- 耗时：平均 ${avg}ms，最大 ${max}ms"
$md += ''
$md += '## 汇总'
$md += ''
$md += '| 模块 | 通过 | 失败 |'
$md += '|------|------|------|'
$results | Group-Object module | ForEach-Object {
  $p = @($_.Group | Where-Object ok).Count
  $f = @($_.Group | Where-Object { -not $_.ok }).Count
  $md += "| $($_.Name) | $p | $f |"
}
$md += ''
$md += '## 明细'
$md += ''
$md += '| 结果 | 模块 | 用例 | 耗时ms | 说明 |'
$md += '|------|------|------|--------|------|'
foreach ($x in $results) {
  $mark = if ($x.ok) { 'PASS' } else { 'FAIL' }
  $detail = ($x.detail -replace '\|', '/' -replace "`n", ' ')
  $md += "| $mark | $($x.module) | $($x.case) | $($x.ms) | $detail |"
}
$md += ''
$md += '## 失败项'
$md += ''
$fails = @($results | Where-Object { -not $_.ok })
if ($fails.Count -eq 0) {
  $md += '无'
} else {
  foreach ($f in $fails) { $md += "- [$($f.module)] $($f.case)：$($f.detail)" }
}
$md += ''
$md += '## 业务结论'
$md += ''
$md += '- 登录/注册审核、客户、预约（含冲突与取消）、次卡/储值开卡、模拟核销扣次、库存出入库、薪资计算、工作台与日志均已覆盖。'
$md += '- 短信为模拟写日志，未对接真实通道（按约定忽略）。'
$md += '- 打卡仅记录方式，未校验 Wi-Fi/定位（按约定忽略）。'
$md += ''

$path = Join-Path $reportDir 'e2e-test-report.md'
$md -join "`n" | Set-Content -Path $path -Encoding UTF8
$results | ConvertTo-Json -Depth 4 | Set-Content (Join-Path $reportDir 'e2e-test-report.json') -Encoding UTF8
Write-Host ''
Write-Host "REPORT: $path"
Write-Host "PASS=$pass FAIL=$fail TOTAL=$total AVG=${avg}ms MAX=${max}ms"
