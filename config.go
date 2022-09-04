package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
)

type Config struct {
  NotesDir string `yaml:"notes_dir"`
  EncDir string `yaml:"encrypted_dir"`
  GitDir string `yaml:"git_dir"`
  GPGPublicKey string `yaml:"gpg_public_key"`
  GPGPrivateKey string `yaml:"gpg_private_key"`
  SSHPrivateKey string `yaml:"ssh_private_key"`
}

func findConfig() string {
  homeDir, _ := homedir.Dir()
  homeConfig := filepath.Join(homeDir, ".config", "note-sync", "config.yaml")
  localConfig := "config.yaml"
  var configPath string
  log.Println("Checking ", homeConfig)
  if _, err := os.Stat(homeConfig); err == nil {
      log.Printf("Home config file found\n");
      configPath = homeConfig
  } else
  if _, err := os.Stat(localConfig); err == nil {
      log.Printf("Local config file found\n");
      configPath = localConfig
  }

  return configPath
}

func getConfig(configFilePath string) Config {

	yamlFile, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		log.Printf("==> Error reading config file: %s\n", err)
	}

	var appConfig Config

	err = yaml.Unmarshal(yamlFile, &appConfig)

	if err != nil {
		log.Printf("==> Error parsing config file: %s\n", err)
	}

	return appConfig

}
