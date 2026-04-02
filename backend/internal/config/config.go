package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"certmanager-backend/pkg/crypto"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构体
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Security SecurityConfig `yaml:"security"`
	JWT      JWTConfig      `yaml:"jwt"`
	CORS     CORSConfig     `yaml:"cors"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `yaml:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// DSN 返回 MySQL DSN 连接字符串
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	AESKey string `yaml:"aes_key"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret       string `yaml:"secret"`
	ExpireHours  int    `yaml:"expire_hours"`
	RefreshHours int    `yaml:"refresh_hours"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 解密配置中的敏感信息（在环境变量覆盖之前）
	// if salt := os.Getenv("CERTMANAGER_ENCRYPT_SALT"); salt != "" {
	if err := decryptConfigValues(&cfg, "ffg5NvYs^"); err != nil {
		return nil, fmt.Errorf("failed to decrypt config: %w", err)
	}
	// }

	// 环境变量覆盖（优先级高于配置文件）
	loadEnvOverrides(&cfg)

	GlobalConfig = &cfg
	return &cfg, nil
}

// loadEnvOverrides 从环境变量加载配置覆盖
func loadEnvOverrides(cfg *Config) {
	// 数据库配置
	if v := os.Getenv("CERTMANAGER_DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("CERTMANAGER_DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("CERTMANAGER_DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("CERTMANAGER_DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("CERTMANAGER_DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}

	// Redis 配置
	if v := os.Getenv("CERTMANAGER_REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("CERTMANAGER_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}

	// JWT 配置
	if v := os.Getenv("CERTMANAGER_JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}

	// 安全配置
	if v := os.Getenv("CERTMANAGER_AES_KEY"); v != "" {
		cfg.Security.AESKey = v
	}

	// CORS 配置 - 支持逗号分隔的多个 origin
	if v := os.Getenv("CERTMANAGER_CORS_ORIGINS"); v != "" {
		origins := strings.Split(v, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		cfg.CORS.AllowedOrigins = origins
	}
}

// decryptConfigValues 解密配置中的敏感字段
func decryptConfigValues(cfg *Config, salt string) error {
	// 解密数据库密码
	if crypto.IsEncrypted(cfg.Database.Password) {
		decrypted, err := crypto.DecryptConfig(cfg.Database.Password, salt)
		if err != nil {
			return fmt.Errorf("decrypt database password failed: %w", err)
		}
		cfg.Database.Password = decrypted
	}

	// 解密 Redis 密码
	if crypto.IsEncrypted(cfg.Redis.Password) {
		decrypted, err := crypto.DecryptConfig(cfg.Redis.Password, salt)
		if err != nil {
			return fmt.Errorf("decrypt redis password failed: %w", err)
		}
		cfg.Redis.Password = decrypted
	}

	// 解密 AES Key
	if crypto.IsEncrypted(cfg.Security.AESKey) {
		decrypted, err := crypto.DecryptConfig(cfg.Security.AESKey, salt)
		if err != nil {
			return fmt.Errorf("decrypt aes key failed: %w", err)
		}
		cfg.Security.AESKey = decrypted
	}

	// 解密 JWT Secret
	if crypto.IsEncrypted(cfg.JWT.Secret) {
		decrypted, err := crypto.DecryptConfig(cfg.JWT.Secret, salt)
		if err != nil {
			return fmt.Errorf("decrypt jwt secret failed: %w", err)
		}
		cfg.JWT.Secret = decrypted
	}

	return nil
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	if GlobalConfig == nil {
		panic("config not loaded")
	}
	return GlobalConfig
}
