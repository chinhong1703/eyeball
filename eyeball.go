package main

import (
	"flag"
	"fmt"
	"github.com/jeremywohl/flatten"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type MasterConfig map[string][]map[string]interface{}

func main() {
	args := parseCli()
	log.Printf(`
		Using the following files:
		Environment: %v
		Master Config File: %v
		Application Properties File: %v
		`, args.Env, args.MasterConfigFile, args.AppPropertiesFile)
	masterConfig := MasterConfig{}
	fileContentAsBytes, err := os.ReadFile(args.MasterConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(fileContentAsBytes, &masterConfig); err != nil {
		log.Fatal(err)
	}

	m := make(map[string]interface{})
	fileContentAsBytes, err = os.ReadFile(args.AppPropertiesFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(fileContentAsBytes, &m); err != nil {
		log.Fatal(err)
	}
	applicationProp, err := flatten.Flatten(m, "", flatten.DotStyle)
	if err != nil {
		log.Fatal(err)
	}

	envMasterConfig, err := getByEnv(masterConfig, args.Env)
	if err != nil {
		log.Fatal(err)
	}
	if err = compare(envMasterConfig, applicationProp); err != nil {
		log.Fatal(err)
	}
	println("application properties is verified")
}

func getByEnv(config MasterConfig, env string) ([]map[string]interface{}, error) {
	val, ok := config[env]
	if !ok {
		return nil, fmt.Errorf("required environment not found in master config file")
	}
	return val, nil
}

func compare(masterConfig []map[string]interface{}, appProperties map[string]interface{}) error {

	for _, required := range masterConfig {
		for requiredKey, requiredValue := range required {
			val, ok := appProperties[requiredKey]
			if !ok {
				return fmt.Errorf("required config not found in app properties. missing config=%v", requiredKey)
			}
			if val != requiredValue {
				return fmt.Errorf("app properties value does not match.\n key=%v \n want=%v \n got=%v", requiredKey, requiredValue, val)
			}

		}
	}
	return nil
}

type CmdLineArgs struct {
	MasterConfigFile  string
	AppPropertiesFile string
	Env               string
}

func parseCli() CmdLineArgs {
	var masterCfgFile string
	var appPropertiesFile string
	var env string

	flag.StringVar(&masterCfgFile, "c", "master-config.yml", "The master configuration file to use")
	flag.StringVar(&appPropertiesFile, "f", "application.yml", "The application properties file to verify")
	flag.StringVar(&env, "env", "prod", "Specify the environment properties to check against")
	flag.Parse()

	return CmdLineArgs{
		MasterConfigFile:  masterCfgFile,
		AppPropertiesFile: appPropertiesFile,
		Env:               env,
	}
}
