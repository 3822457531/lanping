# 阿里云通知短信 · 兰瓶模板申请说明

对接文档：[SendSms](https://help.aliyun.com/zh/sms/developer-reference/api-dysmsapi-2017-05-25-sendsms)  
创建模板 API：[CreateSmsTemplate](https://api.aliyun.com/document/Dysmsapi/2017-05-25/CreateSmsTemplate)

## 前置条件

1. 阿里云账号完成**企业/个体户**实名（通知短信需要营业执照报备）
2. 短信服务控制台申请签名：**兰瓶**（审核通过后才能发）
3. 按下列文案创建 **通知类型** 模板，通过后把 `TemplateCode`（如 `SMS_123456789`）填进 `configs/config.yaml`

## 推荐模板文案（复制到控制台）

签名统一用：`兰瓶`（控制台显示为【兰瓶】）

### 1）生日提醒 → `tpl_birthday`

模板类型：通知短信  

```text
尊敬的${name}，生日快乐！兰瓶美业祝您美丽常在，专属生日礼遇欢迎到店领取。
```

变量：`name`（客户姓名）

### 2）消费通知 → `tpl_consume`

```text
尊敬的${name}，您于${time}完成消费${amount}元（${project}），感谢光临兰瓶美业。
```

变量：`name`、`time`、`amount`、`project`

### 3）预约通知 → `tpl_appoint`

```text
尊敬的${name}，您已预约${date} ${time}的${project}，美容师${staff}，请准时到店。如需改约请联系门店。
```

变量：`name`、`date`、`time`、`project`、`staff`

## 配置示例

```yaml
sms:
  enabled: true
  access_key_id: "你的AK"
  access_key_secret: "你的SK"
  sign_name: "兰瓶"
  tpl_birthday: "SMS_xxxxxxxx"
  tpl_consume: "SMS_yyyyyyyy"
  tpl_appoint: "SMS_zzzzzzzz"
```

或用环境变量（推荐生产）：

- `LP_SMS_ENABLED=true`
- `LP_SMS_ACCESS_KEY_ID`
- `LP_SMS_ACCESS_KEY_SECRET`
- `LP_SMS_SIGN_NAME=兰瓶`
- `LP_SMS_TPL_BIRTHDAY` / `LP_SMS_TPL_CONSUME` / `LP_SMS_TPL_APPOINT`

## 系统触发点

| 业务 | 触发时机 |
|------|----------|
| 消费短信 | 核销确认完成（`finishVerify`） |
| 预约短信 | 新建预约成功 |
| 生日短信 | `POST /api/customers/birthday/notify`（店长/管理员） |

发送记录可在管理端「操作日志 / 短信日志」`GET /api/sms/logs` 查看。

## 注意

- `enabled: false` 或未配 AK 时：只写短信日志为「模拟未发」，不调阿里云
- 模板变量名必须与上表一致，否则阿里云会拒发
- 签名【兰瓶】与营业执照主体需一致或完成授权报备
