package config

type Logger struct {
	Level        string `yaml:"level"`
	Prefix       string `yaml:"prefix"`
	Directory    string `yaml:"directory"`
	ShowLine     bool   `yaml:"show_line"`
	ShowFileName bool   `yaml:"show_file_name"`
}
