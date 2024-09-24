# Eyeball (Golang version)
Your automated partner to help you scan (i.e. eyeball check) your application configuration files and compare it against known configuration to prevent an incorrect configuration (e.g. typo configuration) from ruining your friday evening, because real software developers push code to PROD on fridays.

## Requirements
Golang version 1.21

## Installation
Download dependency packages
```shell
go mod tidy
```
Build
```shell
go build -o eyeball eyeball.go
```

## Usage
Let's imagine you have stored various application properties in your project for the different test environments, example:

- src/main/resources/application-prod.yml
- src/main/resources/application-pre-prod.yml
- src/main/resources/application-uat.yaml
- src/main/resources/application-dev.yaml

They differ in some ways, ie. connecting to different infra for the different environments 
```yaml
spring:
    data:
        jpa:
            url: jdbc:mariadb://localhost:3306/db1 
```

Prepare a master configuration file to serve as the basis for verification:

master-config.yml
```yaml
dev: |
 ...
uat: |
 ...
pre-prod: |
 ...
prod: |
    spring.data.jpa.url=jdbc:mariadb://localhost:3306/db1
```

Run eyeball as follows:

- Check app properties files under a directory
```bash
$ eyeball -env prod -dir properties/prod/ -c master-config.yml
```

- Check a single property file
```bash
$ eyeball -env prod -f src/main/resources/application.yml -c master-config.yml
```

### Compare two YAML files
To activate compare mode and compare two YAML files, use the `--diff` argument along with `-f1` and `-f2` to specify the files to compare:

```bash
$ eyeball --diff -f1 path/to/first/file.yaml -f2 path/to/second/file.yaml
```

In this example:
- `--diff` activates the compare mode.
- `-f1` specifies the path to the first YAML file.
- `-f2` specifies the path to the second YAML file.

Make sure to replace `path/to/first/file.yaml` and `path/to/second/file.yaml` with the actual paths to the YAML files you want to compare.