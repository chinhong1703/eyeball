package main

import (
	"flag"
	"fmt"
	"github.com/jeremywohl/flatten"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

type MasterConfig map[string][]map[string]interface{}

func main() {
	args := parseCli()
	log.Printf(`
 Running eyeball with following arguments:
	Environment: %v
	Master Config File: %v
	Application Properties Directory: %v`, args.Env, args.MasterConfigFile, args.AppPropertiesFile)
	println("==============================")
	masterConfig := MasterConfig{}
	fileContentAsBytes, err := os.ReadFile(args.MasterConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(fileContentAsBytes, &masterConfig); err != nil {
		log.Fatal(err)
	}

	envMasterConfig, err := getByEnv(masterConfig, args.Env)
	if err != nil {
		log.Fatal(err)
	}
	appFiles := getAllYamlFilesInDirectory(args.AppPropertiesFile)
	appProperties := getApplicationPropsFromYaml(appFiles)
	hasErrors := false
	for fileName, applicationProp := range appProperties {
		fmt.Printf("checking file: %v\n", fileName)
		if err = compare(envMasterConfig, applicationProp); err != nil {
			fmt.Println(err.Error())
			hasErrors = true
		} else {
			println("SUCCESS")
		}
		println("==============================")
	}
	println("all application properties are verified")
	println("END")
	if hasErrors {
		os.Exit(1)
	}
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
				continue
			}
			if val != requiredValue {
				return fmt.Errorf("value does not match\n key=%v \n want=%v \n got=%v", requiredKey, requiredValue, val)
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

func getAllYamlFilesInDirectory(dirName string) []string {
	files, err := os.ReadDir(dirName)
	if err != nil {
		panic(err)
	}

	yamlFiles := make([]string, 0)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".yaml" || filepath.Ext(file.Name()) == ".yml" {
			yamlFiles = append(yamlFiles, filepath.Join(dirName, file.Name()))
		}
	}

	return yamlFiles
}

func getApplicationPropsFromYaml(yamlFiles []string) map[string]map[string]interface{} {
	appPropsMap := make(map[string]map[string]interface{})

	for _, file := range yamlFiles {
		m := make(map[string]interface{})

		fileContentAsBytes, err := os.ReadFile(file)
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
		appPropsMap[file] = applicationProp
	}

	return appPropsMap
}
