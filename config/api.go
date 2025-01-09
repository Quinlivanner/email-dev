package config

type API struct {
	EmailCountPerPage  int    `yaml:"email_count_per_page"`
	BaseUrlDomain      string `yaml:"base_url_domain"`
	ShortUrlDomain     string `yaml:"short_url_domain"`
	ShortUrlCodeLength int    `yaml:"short_url_code_length"`
}
