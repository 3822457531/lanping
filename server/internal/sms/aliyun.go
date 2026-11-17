package sms

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

// Config 阿里云短信服务（Dysmsapi）配置。
// 签名、模板需在控制台审核通过后再填 TemplateCode。
type Config struct {
	Enabled         bool
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string // 默认 dysmsapi.aliyuncs.com
	SignName        string // 如：兰瓶
	TplBirthday     string // 生日提醒模板 Code，如 SMS_xxxx
	TplConsume      string // 消费通知模板 Code
	TplAppoint      string // 预约通知模板 Code
}

type Client struct {
	mu     sync.RWMutex
	cfg    Config
	client *dysmsapi.Client
}

func New(cfg Config) (*Client, error) {
	c := &Client{}
	if err := c.Reconfigure(cfg); err != nil {
		return nil, err
	}
	return c, nil
}

// Reconfigure 热更新配置（管理后台保存后调用，无需重启）。
func (c *Client) Reconfigure(cfg Config) error {
	if c == nil {
		return fmt.Errorf("sms client nil")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "dysmsapi.aliyuncs.com"
	}
	if cfg.SignName == "" {
		cfg.SignName = "兰瓶"
	}

	var cli *dysmsapi.Client
	if cfg.Enabled && strings.TrimSpace(cfg.AccessKeyID) != "" && strings.TrimSpace(cfg.AccessKeySecret) != "" {
		created, err := dysmsapi.NewClient(&openapi.Config{
			AccessKeyId:     tea.String(cfg.AccessKeyID),
			AccessKeySecret: tea.String(cfg.AccessKeySecret),
			Endpoint:        tea.String(cfg.Endpoint),
		})
		if err != nil {
			return err
		}
		cli = created
	}

	c.mu.Lock()
	c.cfg = cfg
	c.client = cli
	c.mu.Unlock()
	return nil
}

func (c *Client) snapshot() (Config, *dysmsapi.Client) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cfg, c.client
}

type SendResult struct {
	OK      bool
	BizID   string
	Code    string
	Message string
	Mock    bool
}

func (c *Client) Enabled() bool {
	if c == nil {
		return false
	}
	cfg, cli := c.snapshot()
	return cfg.Enabled && cli != nil
}

func (c *Client) SignName() string {
	if c == nil {
		return ""
	}
	cfg, _ := c.snapshot()
	return cfg.SignName
}

func (c *Client) Config() Config {
	if c == nil {
		return Config{}
	}
	cfg, _ := c.snapshot()
	return cfg
}

func (c *Client) send(phone, templateCode string, params map[string]string) (SendResult, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return SendResult{}, fmt.Errorf("手机号为空")
	}
	if templateCode == "" {
		return SendResult{}, fmt.Errorf("未配置短信模板 Code")
	}
	cfg, cli := c.snapshot()
	if !(cfg.Enabled && cli != nil) {
		raw, _ := json.Marshal(params)
		return SendResult{
			OK:      true,
			Mock:    true,
			Code:    "MOCK",
			Message: fmt.Sprintf("模拟发送【%s】%s %s", cfg.SignName, templateCode, string(raw)),
		}, nil
	}
	paramJSON := "{}"
	if len(params) > 0 {
		b, err := json.Marshal(params)
		if err != nil {
			return SendResult{}, err
		}
		paramJSON = string(b)
	}
	req := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(cfg.SignName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(paramJSON),
	}
	resp, err := cli.SendSms(req)
	if err != nil {
		return SendResult{OK: false, Message: err.Error()}, err
	}
	code := tea.StringValue(resp.Body.Code)
	msg := tea.StringValue(resp.Body.Message)
	biz := tea.StringValue(resp.Body.BizId)
	ok := strings.EqualFold(code, "OK")
	out := SendResult{OK: ok, BizID: biz, Code: code, Message: msg}
	if !ok {
		return out, fmt.Errorf("短信发送失败: %s %s", code, msg)
	}
	return out, nil
}

// SendBirthday 生日祝福。模板变量建议：name
func (c *Client) SendBirthday(phone, name string) (SendResult, error) {
	cfg, _ := c.snapshot()
	return c.send(phone, cfg.TplBirthday, map[string]string{
		"name": name,
	})
}

// SendConsume 消费/核销通知。模板变量建议：name,time,project,amount
func (c *Client) SendConsume(phone, name, timeStr, project, amount string) (SendResult, error) {
	cfg, _ := c.snapshot()
	return c.send(phone, cfg.TplConsume, map[string]string{
		"name":    name,
		"time":    timeStr,
		"project": project,
		"amount":  amount,
	})
}

// SendAppoint 预约通知。模板变量建议：name,date,time,project,staff
func (c *Client) SendAppoint(phone, name, date, timeStr, project, staff string) (SendResult, error) {
	cfg, _ := c.snapshot()
	return c.send(phone, cfg.TplAppoint, map[string]string{
		"name":    name,
		"date":    date,
		"time":    timeStr,
		"project": project,
		"staff":   staff,
	})
}
