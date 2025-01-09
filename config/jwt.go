package config

type Jwt struct {
	SecretKey   string `yaml:"secret_key"`
	ExpiredTime int    `yaml:"expired_time"`
}
