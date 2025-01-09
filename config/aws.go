package config

type AWS struct {
	S3Bucket       string `yaml:"s3_bucket"`
	SQSUrl         string `yaml:"sqs_url"`
	MaxFileSize    int64  `yaml:"max_file_size"`
	FileExpireTime int64  `yaml:"file_expire_time"`
	ConfigRegion   string `yaml:"config_region"`
	ConfigProfile  string `yaml:"config_profile"`
	SmtpHost       string `yaml:"smtp_host"`
	SmtpPort       int    `yaml:"smtp_port"`
	SmtpUsername   string `yaml:"smtp_username"`
	SmtpPassword   string `yaml:"smtp_password"`
}
