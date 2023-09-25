package main

import (
	"flag"
	"fmt"
	"github.com/jeremywohl/flatten"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type MasterConfig map[string][]map[string]interface{}

var args = parseCli()
var masterConfig = MasterConfig{}
var appProperties map[string]map[string]interface{}

func main() {
	logArgs()
	err := readMasterConfig(args.MasterConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	envMasterConfig, err := getByEnv(masterConfig, args.Env)
	if err != nil {
		log.Fatal(err)
	}
	hasErrors := false
	if args.AppPropertiesDir != "" {
		appFiles := getAllYamlFilesInDirectory(args.AppPropertiesDir)
		appProperties = getApplicationPropsFromYaml(appFiles)
	}

	if args.AppPropertiesFile != "" {
		appProperties = getApplicationPropsFromYaml([]string{args.AppPropertiesFile})
	}

	for fileName, applicationProp := range appProperties {
		fmt.Printf(">> checking file: %v\n", fileName)
		if err = compare(envMasterConfig, applicationProp); err != nil {
			fmt.Println(err.Error())
			hasErrors = true
		} else {
			println(">>> SUCCESS")
		}
		println("==============================")
	}

	println(">>>> All application properties are verified")
	println("~~~END~~~")
	if hasErrors {
		os.Exit(1)
	}
}

func readMasterConfig(masterConfigFilename string) error {
	fileContentAsBytes, err := os.ReadFile(masterConfigFilename)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(fileContentAsBytes, &masterConfig); err != nil {
		return err
	}
	return nil
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
	AppPropertiesDir  string
	AppPropertiesFile string
	Env               string
}

func parseCli() CmdLineArgs {
	var masterCfgFile string
	var appPropertiesDir string
	var appPropertiesFile string
	var env string

	flag.StringVar(&masterCfgFile, "c", "master-config.yml", "The master configuration file to use")
	flag.StringVar(&appPropertiesDir, "dir", "", "The directory containing application properties files to verify")
	flag.StringVar(&appPropertiesFile, "f", "", "The application properties file to verify")
	flag.StringVar(&env, "env", "prod", "Specify the environment properties to check against")
	flag.Parse()

	if appPropertiesDir != "" && appPropertiesFile != "" {
		log.Fatal("Cannot handle both -dir and -f arguments. Please use either -dir or -f only.")
	}

	return CmdLineArgs{
		MasterConfigFile:  masterCfgFile,
		AppPropertiesDir:  appPropertiesDir,
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

func logArgs() {

	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintln("> Running eyeball with following arguments:"))
	builder.WriteString(fmt.Sprintf("> Environment: %v\n", args.Env))
	builder.WriteString(fmt.Sprintf("> Master Config File: %v\n", args.MasterConfigFile))

	if args.AppPropertiesDir != "" {
		builder.WriteString(fmt.Sprintf("> Application Properties Directory: %v\n", args.AppPropertiesDir))
	}
	if args.AppPropertiesFile != "" {
		builder.WriteString(fmt.Sprintf("> Application Properties File: %v\n", args.AppPropertiesFile))
	}
	fmt.Print(builder.String())
	println("==============================")
}
