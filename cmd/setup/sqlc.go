package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"
)

func sqlcGenerate() {
	updateSqlcConfig()

	cmdName := "sqlc"
	args := []string{"generate"}

	cmd := exec.Command(cmdName, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("Output:\n%s\n", output)
}

type sqlcConfig struct {
	Version string              `yaml:"version"`
	Sql     []sqlInstanceConfig `yaml:"sql"`
}
type sqlInstanceConfig struct {
	Engine  string    `yaml:"engine"`
	Queries string    `yaml:"queries"`
	Schema  string    `yaml:"schema"`
	Gen     genConfig `yaml:"gen"`
}
type genConfig struct {
	Go goLangConfig `yaml:"go"`
}
type goLangConfig struct {
	Package    string `yaml:"package"`
	Out        string `yaml:"out"`
	SqlPackage string `yaml:"sql_package"`
}

func updateSqlcConfig() {

	entries, err := os.ReadDir("services/")
	if err != nil {
		log.Fatal(err)
	}

	var sqlConfigs []sqlInstanceConfig
	for _, service := range entries {
		name := service.Name()

		sqlConfigs = append(sqlConfigs, sqlInstanceConfig{
			Engine:  "postgresql",
			Queries: fmt.Sprintf("services/%s/%smodels/sql/query.sql", name, name),
			Schema:  fmt.Sprintf("services/%s/%smodels/sql/schema.sql", name, name),
			Gen: genConfig{
				Go: goLangConfig{
					Package:    fmt.Sprintf("%smodels", name),
					Out:        fmt.Sprintf("services/%s/%smodels", name, name),
					SqlPackage: "pgx/v5",
				},
			},
		})
	}

	sc := sqlcConfig{
		Version: "2",
		Sql:     sqlConfigs,
	}

	yamlData, err := yaml.Marshal(&sc)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create("sqlc.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Write the string "hello world" to the file.
	_, err = file.WriteString(string(yamlData))
	if err != nil {
		log.Fatal(err)
	}

}
