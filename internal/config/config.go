package config

import (
	"encoding/json"
	"os"
	"path"
)

const fileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFilePath() string {
	gp := os.Getenv("GOPATH")
	ap := path.Join(gp, fileName)
	return ap
}

func Read() (Config, error) {
	var c Config
	ap := getConfigFilePath()
	b, err := os.ReadFile(ap)
	if err != nil {
		return c, err
	}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	ap := getConfigFilePath()
	err = os.WriteFile(ap, b, 0777)
	if err != nil {
		return err
	}
	return nil
}
