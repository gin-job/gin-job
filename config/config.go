package config

import (
	"os"

	"gorm.io/gorm"
)

type GinJobAuth struct {
	Username string
	Password string
}

type GinJobGorm struct {
	DSN    string
	Config *gorm.Config
}

type GinJobConfig struct {
	TemplatePath string
	Auth         GinJobAuth
	Port         string
	Gorm         GinJobGorm
}

func DefaultConfig() *GinJobConfig {
	templatePath := os.Getenv("TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "../../templates/*"
	}
	gormConfig := &gorm.Config{}
	dsn := "root:gin-job@tcp(localhost:3306)/gin_job?charset=utf8mb4&parseTime=True&loc=Local"
	return &GinJobConfig{
		Port: ":8080",
		Gorm: GinJobGorm{
			DSN:    dsn,
			Config: gormConfig,
		},
		TemplatePath: templatePath,
		Auth: GinJobAuth{
			Username: "admin",
			Password: "gin-job",
		},
	}
}
