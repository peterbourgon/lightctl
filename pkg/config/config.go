package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FilePath is the location of the config file.
var FilePath = func() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "lightctl", "lightctl.conf")
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".lightctl", "lightctl.conf")
	}
	panic("unable to deduce user config dir or user home dir")
}()

type Config struct {
	Username string `json:"username"`
	PSK      string `json:"psk"`
}

func Load() (c Config, err error) {
	buf, err := ioutil.ReadFile(FilePath)
	if err != nil {
		return c, err
	}

	if err := json.Unmarshal(buf, &c); err != nil {
		return c, err
	}

	return c, nil
}
