package Inits

import (
	"email/config"
	"email/global"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// 读取 settings.yaml 配置文件
func InitConfig() {
	const ConfigFilePath = "settings.yaml"
	c := &config.Config{}
	yamlConfig, err := ioutil.ReadFile(ConfigFilePath)
	if err != nil {
		panic(fmt.Errorf("读取配置文件失败: %s", err))
	}
	err = yaml.Unmarshal(yamlConfig, &c)
	if err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}
	global.Config = c
}
