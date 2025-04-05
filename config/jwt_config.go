package config

import (
	"errors"
	"os"
	"time"
)

type JWTConfig struct {
	SecretKey    []byte
	Issuer       string
	TokenExpiry  time.Duration
}

// GetJWTConfig 获取JWT配置
func GetJWTConfig() (*JWTConfig, error) {
	// 从环境变量中读取密钥，优先级高
	secretKey := os.Getenv("JWT_SECRET_KEY")

	if secretKey == "" {
		env := os.Getenv("APP_ENV")
		if env != "production" && env != "staging" {
			secretKey = "dev_secure_key_change_in_production_32chars"
		} else {
			return nil, errors.New("JWT_SECRET_KEY environment variable is not set")
		}
	}

	tokenExpiry := 30 * time.Minute
	if os.Getenv("JWT_EXPIRY_MINUTES") != "" {
		if minutes, err := time.ParseDuration(os.Getenv("JWT_EXPIRY_MINUTES") + "m"); err == nil {
			tokenExpiry = minutes
		}
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "teamtick-backend"
	}

	return &JWTConfig{
		SecretKey:    []byte(secretKey),
		Issuer:       issuer,
		TokenExpiry:  tokenExpiry,
	},nil
}

