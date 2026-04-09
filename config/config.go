package config

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
	return &GinJobConfig{
		Port:         ":8080",
		TemplatePath: "../../templates/",
		Auth: GinJobAuth{
			Username: "admin",
			Password: "gin-job",
		},
	}
}
