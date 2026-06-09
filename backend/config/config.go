package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AppEnv       string `mapstructure:"APP_ENV"`
	ServerPort   string `mapstructure:"SERVER_PORT"`
	DBHost       string `mapstructure:"DB_HOST"`
	DBPort       string `mapstructure:"DB_PORT"`
	DBName       string `mapstructure:"DB_NAME"`
	DBUser       string `mapstructure:"DB_USER"`
	DBPassword   string `mapstructure:"DB_PASSWORD"`
	DBSSLMode    string `mapstructure:"DB_SSLMODE"`
	RedisAddr    string `mapstructure:"REDIS_ADDR"`
	RedisPass    string `mapstructure:"REDIS_PASSWORD"`
	RedisEnabled bool   // true when REDIS_ADDR is set
	StorageDriver   string `mapstructure:"STORAGE_DRIVER"`
	StoragePath     string `mapstructure:"STORAGE_PATH"`
	MinioEndpoint   string `mapstructure:"MINIO_ENDPOINT"`
	MinioAccessKey  string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey  string `mapstructure:"MINIO_SECRET_KEY"`
	MinioBucket     string `mapstructure:"MINIO_BUCKET"`
	MinioUseSSL     bool   `mapstructure:"MINIO_USE_SSL"`
	JWTAccessSecret  string `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
	OnlyOfficeDSURL       string `mapstructure:"ONLYOFFICE_DS_URL"`
	OnlyOfficeJWTSecret   string `mapstructure:"ONLYOFFICE_JWT_SECRET"`
	OnlyOfficeCallbackURL string `mapstructure:"ONLYOFFICE_CALLBACK_URL"`
	OnlyOfficeEnabled     bool   // true when OO URL is set
}

func loadDotEnv(path string) map[string]interface{} {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	result := make(map[string]interface{})
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[0] == val[len(val)-1] {
			val = val[1 : len(val)-1]
		}
		// Keep original key case for Viper matching
		result[key] = val
	}
	return result
}

// allEnvKeys lists every config key that can come from the environment.
// Viper's AutomaticEnv does NOT feed into Unmarshal, so we must Set() each
// one explicitly for Docker/container deployments where no .env file exists.
var allEnvKeys = []string{
	"APP_ENV", "SERVER_PORT",
	"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSLMODE",
	"REDIS_ADDR", "REDIS_PASSWORD",
	"STORAGE_DRIVER", "STORAGE_PATH",
	"MINIO_ENDPOINT", "MINIO_ACCESS_KEY", "MINIO_SECRET_KEY", "MINIO_BUCKET",
	"JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET",
	"ONLYOFFICE_DS_URL", "ONLYOFFICE_JWT_SECRET", "ONLYOFFICE_CALLBACK_URL",
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("APP_ENV", "development")
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_NAME", "file_sys")
	v.SetDefault("DB_USER", "file_sys")
	v.SetDefault("DB_PASSWORD", "file_sys")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("STORAGE_DRIVER", "local")
	v.SetDefault("STORAGE_PATH", "./storage")
	v.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	v.SetDefault("MINIO_ACCESS_KEY", "minioadmin")
	v.SetDefault("MINIO_SECRET_KEY", "minioadmin")
	v.SetDefault("MINIO_BUCKET", "file-sys")

	// 1) .env file (lowest priority among explicit sources)
	envMap := loadDotEnv(".env")
	if envMap == nil {
		envMap = loadDotEnv("../.env")
	}
	if envMap != nil {
		for k, val := range envMap {
			v.Set(k, val)
		}
	}

	// 2) Environment variables override .env (required for Docker; Viper's
	//    AutomaticEnv doesn't feed Unmarshal so we Set them manually)
	for _, key := range allEnvKeys {
		if val := os.Getenv(key); val != "" {
			v.Set(key, val)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	if cfg.RedisAddr != "" && cfg.RedisAddr != "localhost:6379" {
		cfg.RedisEnabled = true
	}
	if cfg.OnlyOfficeDSURL != "" {
		cfg.OnlyOfficeEnabled = true
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword +
		"@" + c.DBHost + ":" + c.DBPort +
		"/" + c.DBName + "?sslmode=" + c.DBSSLMode
}
