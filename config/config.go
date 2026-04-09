package config

import "os"

type GinJobAuth struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

type GinJobConfig struct {
	TemplatePath string     `json:"template_path" yaml:"template_path"`
	Auth         GinJobAuth `json:"auth" yaml:"auth"`
	Port         string     `json:"port" yaml:"port"`
}

func DefaultConfig() *GinJobConfig {
	templatePath := os.Getenv("TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "../../templates/*"
	}
	return &GinJobConfig{
		Port:         ":8080",
		TemplatePath: templatePath,
		Auth: GinJobAuth{
			Username: "admin",
			Password: "gin-job",
		},
	}
}
