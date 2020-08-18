package main

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/kelseyhightower/envconfig"
	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v2"
	"os"
)

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func readFile(configFilePath string, cfg *Config) {
	f, err := os.Open(configFilePath)
	if err != nil {
		processError(err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		processError(err)
	}
}

func parseConfig() Config {
	var cfg = Config{}
	defaults.SetDefaults(&cfg)
	if len(os.Args) > 1 {
		var configFile = os.Args[1]
		readFile(configFile, &cfg)
	}
	readEnv(&cfg)
	var validate = validator.New()
	var err = validate.Struct(&cfg)

	if err != nil {
		panic(err.Error())
	}
	return cfg
}

type Config struct {
	Server struct {
		Port string `default:"8080", validate:"required", yaml:"port", envconfig:"SERVER_PORT"`
		SSL  struct {
			Enable   bool   `default:"false", validate:"required", yaml:"enable", envconfig:"SERVER_SSL_ENABLE"`
			CertPath string `yaml:"certPath", envconfig:"SERVER_SSL_CERT_PATH"`
			KeyPath  string `yaml:"keyPath", envconfig:"SERVER_SSL_KEY_PATH"`
		} `yaml:"ssl"`
	} `yaml:"server"`
	App struct {
		OverrideVolumeCollisions bool   `default:"false", validate:"required", yaml:"overrideVolumeCollisions", envconfig:"APP_OVERRIDE_VOLUME_COLLISIONS"`
		TargetContainerName      string `validate:"required", yaml:"targetContainerName", envconfig:"APP_TARGET_CONTAINER_NAME"`
		EmptyDir                 struct {
			VolumeName string `validate:"required", yaml:"volumeName", envconfig:"APP_EMPTYDIR_VOLUME_NAME"`
			MountPath  string `validate:"required", yaml:"mountPath", envconfig:"APP_EMPTYDIR_MOUNT_PATH"`
		} `yaml:"emptydir"`
	} `yaml:"app"`
}