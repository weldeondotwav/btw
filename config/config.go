package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type AppConfig struct {
	RemindersFilePath string `json:"reminders_file_path,omitempty"`
}

func NewDefaultConfig() AppConfig {

	AppDataPath := path.Clean(os.Getenv("APPDATA"))

	configPath := AppDataPath + "\\btw_reminders" + "\\reminders.txt"

	return AppConfig{
		RemindersFilePath: configPath,
	}
}

func (c *AppConfig) Save() error {
	fmt.Println("Saving config...")
	AppDataPath := path.Clean(os.Getenv("APPDATA"))
	configPath := AppDataPath + "\\btw_reminders" + "\\config.json"

	configBytes, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	os.MkdirAll(AppDataPath+"\\btw_reminders", 0666)

	err = os.WriteFile(configPath, configBytes, 0666)

	if err != nil {
		return err
	}

	fmt.Printf("Wrote config to %s\n", configPath)

	return nil
}

// Loads the config as long as its in the normal spot
func Read() (*AppConfig, error) {
	AppDataPath := path.Clean(os.Getenv("APPDATA"))
	configPath := AppDataPath + "\\btw_reminders" + "\\config.json"
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	out := &AppConfig{}

	err = json.Unmarshal(data, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func Path() string {
	AppDataPath := path.Clean(os.Getenv("APPDATA"))
	configPath := AppDataPath + "\\btw_reminders" + "\\config.json"

	return configPath
}
