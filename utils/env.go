package utils

import (
	"encoding/json"
	"github.com/aws/jsii-runtime-go"
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

	if _, err := os.Stat(configs[0]); os.IsNotExist(err) {
		panic("Config file does not exist")
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
