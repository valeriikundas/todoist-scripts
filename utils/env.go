package utils

import (
	"encoding/json"
	"github.com/aws/jsii-runtime-go"
	"log"
	"os"
	"strings"
)

// todo: read about project structure
func ReadConfig(configs ...string) *map[string]*string {
	if len(configs) == 0 {
		configs = []string{"./config.json"}
	}

	if len(configs) > 1 {
		panic("Only one config file is allowed")
	}

	_, err := os.Stat(configs[0])
	if err != nil {
		if os.IsNotExist(err) {
			log.Print("Config file not found, using default config")
			return &map[string]*string{
				"ExcludeFromZeroProjectsList": jsii.String(""),
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
