package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type (
	Config struct {
		HTTP          HTTPConfig     `yaml:"http"`
		SMTP          SMTPConfig     `yaml:"smtp"`
		Postgres      PostgresConfig `yaml:"pg"`
		AuthConfig    AuthConfig     `yaml:"auth"`
		Env           string         `yaml:"env"`
		MigrationPath string         `yaml:"migrationPath"`
	}

	HTTPConfig struct {
		Host               string        `yaml:"host"`
		Port               string        `yaml:"port"`
		ReadTimeOut        time.Duration `yaml:"readTimeout"`
		WriteTimeOut       time.Duration `yaml:"writeTimeOut"`
		MaxHeaderMegabytes int           `yaml:"MaxHeaderMegabytes"`
	}

	SMTPConfig struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `env:"SMTP_PASSWORD"`
	}

	PostgresConfig struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `env:"POSTGRES_PASSWORD"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	}

	AuthConfig struct {
		JWT                    JWTConfig `yaml:"jwt"`
		PasswordSalt           string    `env:"PASSWORD_SALT"`
		VerificationCodeLength int       `yaml:"verificationCodeLength"`
	}

	JWTConfig struct {
		AccessTokenTTL  time.Duration `yaml:"accessTTL"`
		RefreshTokenTTL time.Duration `yaml:"refreshTTL"`
		SignedKey       string        `env:"JWT_SIGNED_KEY"`
	}
)

func InitConfig(configPath string) *Config {
	if configPath == "" {
		log.Fatal("configPath is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("cannot find config file.. \"%s\"", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config file: \"%s\"", err.Error())
	}

	return &cfg
}

//func Init(configDir string, configFile string) (*Config, error) {
//	if err := parseConfig(configDir, configFile); err != nil {
//		return nil, err
//	}
//
//	readTimeoutDuration, err := time.ParseDuration(viper.GetString("http.readTimeout"))
//	if err != nil {
//		return nil, err
//	}
//
//	writeTimeoutDuration, err := time.ParseDuration(viper.GetString("http.writeTimeout"))
//	if err != nil {
//		return nil, err
//	}
//
//	accessTokenTTL, err := time.ParseDuration(viper.GetString("auth.jwt.accessTTL"))
//	if err != nil {
//		return nil, err
//	}
//	refreshTokenTTL, err := time.ParseDuration(viper.GetString("auth.jwt.refreshTTL"))
//	if err != nil {
//		return nil, err
//	}
//
//	return &Config{
//		HTTP: HTTPConfig{
//			Host:               viper.GetString("http.host"),
//			Port:               viper.GetString("http.port"),
//			ReadTimeOut:        readTimeoutDuration,
//			WriteTimeOut:       writeTimeoutDuration,
//			MaxHeaderMegabytes: viper.GetInt("http.maxHeaderBytes"),
//		},
//		SMTP: SMTPConfig{
//			Host:     viper.GetString("smtp.host"),
//			Port:     viper.GetInt("smtp.port"),
//			User:     viper.GetString("smtp.user"),
//			Password: os.Getenv("SMTP_PASSWORD"),
//		},
//		Postgres: PostgresConfig{
//			Host:     viper.GetString("pg.host"),
//			Port:     viper.GetString("pg.port"),
//			User:     viper.GetString("pg.user"),
//			Password: os.Getenv("DB_PASSWORD"),
//			DBName:   viper.GetString("pg.dbname"),
//			SSLMode:  viper.GetString("pg.sslmode"),
//		},
//		AuthConfig: AuthConfig{
//			JWT: JWTConfig{
//				AccessTokenTTL:  accessTokenTTL,
//				RefreshTokenTTL: refreshTokenTTL,
//				SignedKey:       os.Getenv("JWT_SIGNED_KEY"),
//			},
//			PasswordSalt:           os.Getenv("PASSWORD_SALT"),
//			VerificationCodeLength: viper.GetInt("auth.verificationCodeLength"),
//		},
//		Env: viper.GetString("env"),
//	}, nil
//}
//
//func parseConfig(configDir string, configFile string) error {
//	viper.AddConfigPath(configDir)
//	viper.SetConfigName(configFile)
//
//	if err := viper.ReadInConfig(); err != nil {
//		return err
//	}
//
//	if err := godotenv.Load(); err != nil {
//		return err
//	}
//
//	return viper.MergeInConfig()
//}
