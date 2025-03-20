package cfg

import (
	"gopkg.in/yaml.v2"
	"note/shell"
	"os"
)

type Config struct {
	App struct {
		Db     string `yaml:"db"`
		Editor string `yaml:"editor"`
	} `yaml:"app"`
	Git struct {
		RemoteURL string `yaml:"url"`
		User      string `yaml:"user"`
		Password  string `yaml:"password"`
		Branch    string `yaml:"branch"`
	} `yaml:"github"`
}

var DefaultCfg = &Config{}

var Path = "/etc/note_config.yaml"

func init() {
	//flag.StringVar(&Path, "config", "", "config file path")
	//flag.Parse()
	//log.Log("ConfigPath:", Path)
	loadConfig()
}

func loadConfig() {
	data, err := os.ReadFile(Path)
	if err != nil {
		shell.Log(err)
		return
	}
	err = yaml.Unmarshal(data, &DefaultCfg)
	if err != nil {
		shell.Log("config 解析失败:", err)
		return
	}
}
