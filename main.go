package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	invv1alpha1 "github.com/sdcio/config-server/apis/inv/v1alpha1"
	"github.com/sdcio/config-server/pkg/git/auth"
	schemaloader "github.com/sdcio/config-server/pkg/schema"
	"github.com/sdcio/schema-server/pkg/config"
	"github.com/sdcio/schema-server/pkg/schema"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"
)

const (
	tmpPath     = "tmp"
	schemasPath = "schemas"
)

func main() {
	args := os.Args

	os.MkdirAll(tmpPath, 0755|os.ModeDir)
	os.MkdirAll(schemasPath, 0755|os.ModeDir)

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.DebugLevel)

	if len(args) < 1 {
		panic("cannot execute need config and base dir")
	}

	// schemaconfig is the first arg
	schemacr, err := getConfig(args[1])
	if err != nil {
		panic(err)
	}

	/*
		basepath := args[2]
		if err := validatePath(basepath); err != nil {
			panic(err)
		}
	*/

	schemaLoader, err := schemaloader.NewLoader(
		filepath.Join(tmpPath),
		filepath.Join(schemasPath),
		NewNopResolver(),
	)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	schemaspec := &schemacr.Spec
	schemaLoader.AddRef(ctx, schemaspec)
	_, dirExists, err := schemaLoader.GetRef(ctx, schemaspec.GetKey())
	if err != nil {
		panic(err)
	}
	if !dirExists {
		fmt.Println("loading...")
		if err := schemaLoader.Load(ctx, schemaspec.GetKey(), types.NamespacedName{
			Namespace: schemacr.Namespace,
			Name:      schemaspec.Credentials,
		}); err != nil {
			panic(err)
		}
	}

	//models := initSlice(schemacr.Spec.Schema.Models, ".")
	//includes := initSlice(schemacr.Spec.Schema.Includes, "")
	//excludes := initSlice(schemacr.Spec.Schema.Excludes, "")

	//fmt.Println("models", getNewBase(basepath, models))
	//fmt.Println("includes", getNewBase(basepath, includes))

	schemaConfig := &config.SchemaConfig{
		Name:    "",
		Vendor:  schemacr.Spec.Provider,
		Version: schemacr.Spec.Version,
		//Files:       getNewBase(basepath, models),
		Files: schemaspec.GetNewSchemaBase(schemasPath).Models,
		//Directories: getNewBase(basepath, includes),
		Directories: schemaspec.GetNewSchemaBase(schemasPath).Includes,
		//Excludes:    excludes,
		Excludes: schemaspec.GetNewSchemaBase(schemasPath).Excludes,
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

/*
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
*/

func NewNopResolver() auth.CredentialResolver {
	return &secretNopResolver{}
}

var _ auth.CredentialResolver = &secretNopResolver{}

type secretNopResolver struct{}

func (r *secretNopResolver) ResolveCredential(ctx context.Context, nsn types.NamespacedName) (auth.Credential, error) {
	return nil, nil
}
