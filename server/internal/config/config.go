package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP struct {
		Addr string `yaml:"addr"`
		Mode string `yaml:"mode"`
	} `yaml:"http"`
	DB struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`
	JWT struct {
		Secret      string `yaml:"secret"`
		ExpireHours int    `yaml:"expire_hours"`
	} `yaml:"jwt"`
	Migrate struct {
		Auto bool   `yaml:"auto"`
		Dir  string `yaml:"dir"`
	} `yaml:"migrate"`
	// 阿里云短信服务 Dysmsapi（通知短信：生日/消费/预约）
	SMS struct {
		Enabled         bool   `yaml:"enabled"`
		AccessKeyID     string `yaml:"access_key_id"`
		AccessKeySecret string `yaml:"access_key_secret"`
		Endpoint        string `yaml:"endpoint"`
		SignName        string `yaml:"sign_name"`
		TplBirthday     string `yaml:"tpl_birthday"`
		TplConsume      string `yaml:"tpl_consume"`
		TplAppoint      string `yaml:"tpl_appoint"`
	} `yaml:"sms"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	cfg.HTTP.Addr = ":8080"
	cfg.HTTP.Mode = "debug"
	cfg.DB.DSN = "lanping:lanping@tcp(127.0.0.1:3306)/lanping?parseTime=true&loc=Local&charset=utf8mb4"
	cfg.JWT.Secret = "lanping-dev-secret-change-me"
	cfg.JWT.ExpireHours = 72
	cfg.Migrate.Auto = true
	cfg.Migrate.Dir = "migrations"

	if path != "" {
		b, err := os.ReadFile(path)
		if err == nil {
			_ = yaml.Unmarshal(b, cfg)
		}
	}

	if v := os.Getenv("LP_HTTP_ADDR"); v != "" {
		cfg.HTTP.Addr = v
	}
	if v := os.Getenv("LP_DB_DSN"); v != "" {
		cfg.DB.DSN = v
	}
	if v := os.Getenv("LP_JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("LP_AUTO_MIGRATE"); v != "" {
		cfg.Migrate.Auto = v == "1" || v == "true" || v == "TRUE"
	}
	if v := os.Getenv("LP_JWT_EXPIRE_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.JWT.ExpireHours = n
		}
	}
	if v := os.Getenv("LP_SMS_ENABLED"); v != "" {
		cfg.SMS.Enabled = v == "1" || v == "true" || v == "TRUE"
	}
	if v := os.Getenv("LP_SMS_ACCESS_KEY_ID"); v != "" {
		cfg.SMS.AccessKeyID = v
	}
	if v := os.Getenv("LP_SMS_ACCESS_KEY_SECRET"); v != "" {
		cfg.SMS.AccessKeySecret = v
	}
	if v := os.Getenv("LP_SMS_SIGN_NAME"); v != "" {
		cfg.SMS.SignName = v
	}
	if v := os.Getenv("LP_SMS_TPL_BIRTHDAY"); v != "" {
		cfg.SMS.TplBirthday = v
	}
	if v := os.Getenv("LP_SMS_TPL_CONSUME"); v != "" {
		cfg.SMS.TplConsume = v
	}
	if v := os.Getenv("LP_SMS_TPL_APPOINT"); v != "" {
		cfg.SMS.TplAppoint = v
	}
	if cfg.SMS.SignName == "" {
		cfg.SMS.SignName = "兰瓶"
	}
	if cfg.SMS.Endpoint == "" {
		cfg.SMS.Endpoint = "dysmsapi.aliyuncs.com"
	}
	return cfg, nil
}
