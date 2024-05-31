package main

type AppConfig struct {
	Port           int64  `yaml:"port"`
	ApiKey         string `yaml:"api_key"`
	ApiSecret      string `yaml:"api_secret"`
	Path           string `yaml:"path"`
	Debug          bool   `yaml:"debug"`
	Compress       bool   `yaml:"compress"`
	DeleteEmptyDir bool   `yaml:"delete_empty_dir"`
}
