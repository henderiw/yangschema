package main

import (
	"fmt"
	"os"
	"path/filepath"

	invv1alpha1 "github.com/sdcio/config-server/apis/inv/v1alpha1"
	"github.com/sdcio/schema-server/pkg/config"
	"github.com/sdcio/schema-server/pkg/schema"
	"sigs.k8s.io/yaml"
	log "github.com/sirupsen/logrus"

)

func main() {
	args := os.Args

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.DebugLevel)

	if len(args) < 2 {
		panic("cannot execute need config and base dir")
	}

	// schemaconfig is the first arg
	s, err := getConfig(args[1])
	if err != nil {
		panic(err)
	}

	basepath := args[2]
	if err := validatePath(basepath); err != nil {
		panic(err)
	}

	models := initSlice(s.Spec.Schema.Models, ".")
	includes := initSlice(s.Spec.Schema.Includes, "")
	excludes := initSlice(s.Spec.Schema.Excludes, "")

	fmt.Println("models", getNewBase(basepath, models))
	fmt.Println("includes", getNewBase(basepath, includes))

	schemaConfig := &config.SchemaConfig{
		Name:        "",
		Vendor:      s.Spec.Provider,
		Version:     s.Spec.Version,
		Files:       getNewBase(basepath, models),
		Directories: getNewBase(basepath, includes),
		Excludes:    excludes,
	}

	x, err := schema.NewSchema(schemaConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println(x)

}

func getConfig(schemaConfigPath string) (*invv1alpha1.Schema, error) {
	b, err := os.ReadFile(schemaConfigPath)
	if err != nil {
		panic(err)
	}

	schema := &invv1alpha1.Schema{}
	if err := yaml.Unmarshal(b, schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func validatePath(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("input path needs to be a directory")
	}
	return nil
}

func initSlice(in []string, init string) []string {
	if len(in) == 0 {
		if init != "" {
			return []string{init}
		} else {
			return []string{}
		}
	}
	return in
}

func getNewBase(basePath string, in []string) []string {
	str := make([]string, 0, len(in))
	for _, s := range in {
		str = append(str, filepath.Join(basePath, s))
	}
	return str
}
