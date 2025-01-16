package config

type System struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	Env                bool   `yaml:"env"`
	SmtpMaxConnections int    `yaml:"smtp_max_connections"`
}
