package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/jeremywohl/flatten"
	"gopkg.in/yaml.v3"
)

type MasterConfig map[string][]map[string]interface{}

type MultiEnvConfig struct {
	Environments map[string]map[string]interface{}
}

var args = parseCli()
var masterConfig = MasterConfig{}
var appProperties map[string]map[string]interface{}

func main() {
	logArgs()

	if args.MultiEnvMode {
		if args.MultiEnvFile == "" {
			log.Fatal("Multi-environment file (-mf) must be specified when using -multi mode")
		}
		if args.MasterConfigFile == "" {
			log.Fatal("Master config file (-c) must be specified when using -multi mode")
		}
		if err := handleMultiEnvFile(args.MultiEnvFile); err != nil {
			log.Fatalf("Error processing multi-environment file: %v", err)
		}
		return
	}

	if args.DiffMode {
		if args.File1 == "" || args.File2 == "" {
			log.Fatal("Both -f1 and -f2 arguments must be provided in diff mode")
		}
		if err := compareYamlFiles(args.File1, args.File2); err != nil {
			log.Fatalf("Error comparing YAML files: %v", err)
		}
		return
	}

	if err := readMasterConfig(args.MasterConfigFile); err != nil {
		log.Fatalf("Error reading master config file: %v", err)
	}

	masterConfigForEnv, err := getByEnv(masterConfig, args.Env)
	if err != nil {
		log.Fatalf("Error getting environment config: %v", err)
	}

	if args.AppPropertiesDir != "" {
		yamlFiles := getAllYamlFilesInDirectory(args.AppPropertiesDir)
		appPropsMap := getApplicationPropsFromYaml(yamlFiles)
		for file, appProps := range appPropsMap {
			if err := compare(masterConfigForEnv, appProps); err != nil {
				log.Printf("Mismatch in file %s: %v", file, err)
			} else {
				log.Printf("File %s matches the master config", file)
			}
		}
	} else if args.AppPropertiesFile != "" {
		yamlFiles := []string{args.AppPropertiesFile}
		appPropsMap := getApplicationPropsFromYaml(yamlFiles)
		for file, appProps := range appPropsMap {
			if err := compare(masterConfigForEnv, appProps); err != nil {
				log.Printf("Mismatch in file %s: %v", file, err)
			} else {
				log.Printf("File %s matches the master config", file)
			}
		}
	} else {
		log.Fatal("No application properties file or directory specified")
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
				return fmt.Errorf("required key not found\n key=%v", requiredKey)
			}
			if val != requiredValue {
				return fmt.Errorf("value does not match\n key=%v \n want=%v \n got=%v", requiredKey, requiredValue, val)
			}
		}
	}
	return nil
}

func compareYamlFiles(file1, file2 string) error {
	content1, err := os.ReadFile(file1)
	if err != nil {
		return err
	}

	content2, err := os.ReadFile(file2)
	if err != nil {
		return err
	}

	var yaml1, yaml2 map[string]interface{}
	if err = yaml.Unmarshal(content1, &yaml1); err != nil {
		return err
	}
	if err = yaml.Unmarshal(content2, &yaml2); err != nil {
		return err
	}

	flatYaml1, err := flatten.Flatten(yaml1, "", flatten.DotStyle)
	if err != nil {
		return err
	}
	flatYaml2, err := flatten.Flatten(yaml2, "", flatten.DotStyle)
	if err != nil {
		return err
	}

	identical := true

	for key, value1 := range flatYaml1 {
		if value2, ok := flatYaml2[key]; ok {
			if !reflect.DeepEqual(value1, value2) {
				fmt.Printf("Difference found:\nKey: %s\nFile1: %v\nFile2: %v\n\n", key, value1, value2)
				identical = false
			}
		} else {
			fmt.Printf("Key %s found in File1 but not in File2\n\n", key)
			identical = false
		}
	}

	for key := range flatYaml2 {
		if _, ok := flatYaml1[key]; !ok {
			fmt.Printf("Key %s found in File2 but not in File1\n\n", key)
			identical = false
		}
	}

	if identical {
		fmt.Println("The two YAML files are identical.")
	}

	return nil
}

type CmdLineArgs struct {
	MasterConfigFile  string
	AppPropertiesDir  string
	AppPropertiesFile string
	Env               string
	DiffMode          bool
	File1             string
	File2             string
	MultiEnvMode      bool
	MultiEnvFile      string
}

func parseCli() CmdLineArgs {
	var masterCfgFile string
	var appPropertiesDir string
	var appPropertiesFile string
	var env string
	var diffMode bool
	var file1 string
	var file2 string
	var multiEnvMode bool
	var multiEnvFile string

	flag.StringVar(&masterCfgFile, "c", "master-config.yml", "The master configuration file to use")
	flag.StringVar(&appPropertiesDir, "dir", "", "The directory containing application properties files to verify")
	flag.StringVar(&appPropertiesFile, "f", "", "The application properties file to verify")
	flag.StringVar(&env, "env", "prod", "Specify the environment properties to check against")
	flag.BoolVar(&diffMode, "diff", false, "Activate compare mode")
	flag.StringVar(&file1, "f1", "", "The first YAML file to compare")
	flag.StringVar(&file2, "f2", "", "The second YAML file to compare")
	flag.BoolVar(&multiEnvMode, "multi", false, "Activate multi-environment validation mode")
	flag.StringVar(&multiEnvFile, "mf", "", "The multi-environment YAML file to validate")
	flag.Parse()

	if appPropertiesDir != "" && appPropertiesFile != "" {
		log.Fatal("Cannot handle both -dir and -f arguments. Please use either -dir or -f only.")
	}

	return CmdLineArgs{
		MasterConfigFile:  masterCfgFile,
		AppPropertiesDir:  appPropertiesDir,
		AppPropertiesFile: appPropertiesFile,
		Env:               env,
		DiffMode:          diffMode,
		File1:             file1,
		File2:             file2,
		MultiEnvMode:      multiEnvMode,
		MultiEnvFile:      multiEnvFile,
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
	if args.MultiEnvMode {
		builder.WriteString(fmt.Sprintln("> Running in multi-environment mode"))
		builder.WriteString(fmt.Sprintf("> Multi-environment File: %v\n", args.MultiEnvFile))
		builder.WriteString(fmt.Sprintf("> Master Config File: %v\n", args.MasterConfigFile))
	} else if args.DiffMode {
		builder.WriteString(fmt.Sprintln("> Running in compare mode"))
		builder.WriteString(fmt.Sprintf("> File 1: %v\n", args.File1))
		builder.WriteString(fmt.Sprintf("> File 2: %v\n", args.File2))
	} else {
		builder.WriteString(fmt.Sprintf("> Environment: %v\n", args.Env))
		builder.WriteString(fmt.Sprintf("> Master Config File: %v\n", args.MasterConfigFile))

		if args.AppPropertiesDir != "" {
			builder.WriteString(fmt.Sprintf("> Application Properties Directory: %v\n", args.AppPropertiesDir))
		}
		if args.AppPropertiesFile != "" {
			builder.WriteString(fmt.Sprintf("> Application Properties File: %v\n", args.AppPropertiesFile))
		}
	}

	fmt.Print(builder.String())
	println("==============================")
}

func handleMultiEnvFile(filename string) error {
	if err := readMasterConfig(args.MasterConfigFile); err != nil {
		return fmt.Errorf("error reading master config file: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	var docs []map[string]interface{}
	decoder := yaml.NewDecoder(strings.NewReader(string(content)))
	for {
		var doc map[string]interface{}
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error decoding YAML: %v", err)
		}
		if doc != nil {
			docs = append(docs, doc)
		}
	}

	foundMismatch := false
	for _, doc := range docs {
		if spring, ok := doc["spring"].(map[string]interface{}); ok {
			if config, ok := spring["config"].(map[string]interface{}); ok {
				if activate, ok := config["activate"].(map[string]interface{}); ok {
					if profile, ok := activate["on-profile"].(string); ok {
						flat, err := flatten.Flatten(doc, "", flatten.DotStyle)
						if err != nil {
							return fmt.Errorf("error flattening document: %v", err)
						}

						fmt.Printf("\n=== Environment: %s ===\n", profile)
						fmt.Printf("Configuration keys found: %d\n", len(flat))
						
						masterConfigForEnv, err := getByEnv(masterConfig, profile)
						if err != nil {
							fmt.Printf("WARNING: Environment %s not found in master config\n", profile)
							foundMismatch = true
							continue
						}

						if err := compare(masterConfigForEnv, flat); err != nil {
							fmt.Printf("MISMATCH in %s environment: %v\n", profile, err)
							foundMismatch = true
						} else {
							fmt.Printf("SUCCESS: %s environment matches master config\n", profile)
						}

						// Check for sensitive data in non-prod environments
						if profile != "prod" {
							checkSensitiveData(flat, profile)
						}
					}
				}
			}
		}
	}

	if foundMismatch {
		return fmt.Errorf("one or more environments had mismatches")
	}
	return nil
}

func checkSensitiveData(config map[string]interface{}, env string) {
	sensitivePatterns := []string{
		"password",
		"secret",
		"key",
		"token",
		"credential",
	}

	for key, value := range config {
		for _, pattern := range sensitivePatterns {
			if strings.Contains(strings.ToLower(key), pattern) {
				strValue, ok := value.(string)
				if ok && !strings.Contains(strValue, "${") {
					fmt.Printf("WARNING: Possible hardcoded sensitive data found in %s environment\n"+
						"Key: %s\n"+
						"Consider using environment variables instead\n", env, key)
				}
			}
		}
	}
}

