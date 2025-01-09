package config

type OtherInfo struct {
	TagColor        string   `yaml:"tag_color"`
	JwtTokenSecret  string   `yaml:"jwttoken_secret"`
	UnDetectUrlPath []string `yaml:"undetect_url_path"`
}
