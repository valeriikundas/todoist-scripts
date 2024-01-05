package utils

import (
	"encoding/json"
	"github.com/aws/jsii-runtime-go"
	"log"
	"os"
	"strings"
)

func ReadConfig(configs ...string) *map[string]*string {
	// todo: read about project structure
	// todo: remove config, use env for everything

	if len(configs) == 0 {
		configs = []string{"./config.json"}
	}

	if len(configs) > 1 {
		panic("Only one config file is allowed")
	}

	_, err := os.Stat(configs[0])
	if err != nil {
		if os.IsNotExist(err) {
			log.Print("Config file not found, using environment variables")
			excludeFromZeroProjectsList, ok := os.LookupEnv("ExcludeFromZeroProjectsList")
			if !ok {
				panic("ExcludeFromZeroProjectsList environment variable is missing")
			}
			return &map[string]*string{
				"ExcludeFromZeroProjectsList": jsii.String(excludeFromZeroProjectsList),
			}
		} else {
			panic(err)
		}
	}

	configFileName := configs[0]

	file, err := os.Open(configFileName)
	must(err)

	decoder := json.NewDecoder(file)
	var config struct {
		ExcludeFromZeroProjectsList []string
	}
	err = decoder.Decode(&config)
	must(err)

	zeroProjectsListJoined := strings.Join(config.ExcludeFromZeroProjectsList, ";")
	envVars := &map[string]*string{
		"ExcludeFromZeroProjectsList": jsii.String(zeroProjectsListJoined),
	}
	return envVars
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
