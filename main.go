package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type KubectlConfig struct {
	Kind           string                    `yaml:"kind"`
	ApiVersion     string                    `yaml:"apiVersion"`
	CurrentContext string                    `yaml:"current-context"`
	Clusters       []*KubectlClusterWithName `yaml:"clusters"`
	Contexts       []*KubectlContextWithName `yaml:"contexts"`
	Users          []*KubectlUserWithName    `yaml:"users"`
}

type KubectlUser struct {
	ClientCertificateData string `yaml:"client-certificate-data,omitempty"`
	ClientKeyData         string `yaml:"client-key-data,omitempty"`
	Password              string `yaml:"password,omitempty"`
	Username              string `yaml:"username,omitempty"`
	Token                 string `yaml:"token,omitempty"`
}

type KubectlUserWithName struct {
	Name string      `yaml:"name"`
	User KubectlUser `yaml:"user"`
}

type KubectlContext struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}

type KubectlContextWithName struct {
	Name    string         `yaml:"name"`
	Context KubectlContext `yaml:"context"`
}

type KubectlCluster struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
}

type KubectlClusterWithName struct {
	Name    string         `yaml:"name"`
	Cluster KubectlCluster `yaml:"cluster"`
}

const KUBECONFIG_ENV = "KUBECONFIG"
const KUBECONFIG_ENV_KEY = "$KUBECONFIG"

var KUBECONFIG_DEFAULT_PATH = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}
	return filepath.Join(home, ".kube", "config")
}()

func ParseKubeConfig(path string) (*KubectlConfig, error) {

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var kubeConfig KubectlConfig
	err = yaml.Unmarshal(file, &kubeConfig)
	if err != nil {
		return nil, err
	}

	return &kubeConfig, nil
}

func ValidateOnlyOneContext(toBeAppend KubectlConfig, kubectlConfig KubectlConfig) error {

	if len(toBeAppend.Clusters) != 1 {
		return errors.New("only one cluster can be merged into original kubeconfig")
	}

	if len(toBeAppend.Users) != 1 {
		return errors.New("only one user can be merged into original kubeconfig")
	}

	if len(toBeAppend.Contexts) != 1 {
		return errors.New("only one context can be merged into original kubeconfig")
	}

	return nil
}

func ValidateDuplication(toBeAppend KubectlConfig, kubectlConfig KubectlConfig) error {

	for _, v1 := range kubectlConfig.Clusters {
		for _, v2 := range toBeAppend.Clusters {
			if strings.EqualFold(v1.Name, v2.Name) {
				return fmt.Errorf("a cluster entry with %s already exists in kubeconfig, merge failed", v1.Name)
			}
		}
	}

	for _, v1 := range kubectlConfig.Users {
		for _, v2 := range toBeAppend.Users {
			if strings.EqualFold(v1.Name, v2.Name) {
				return fmt.Errorf("a user entry with %s already exists in kubeconfig, merge failed", v1.Name)
			}
		}
	}

	for _, v1 := range kubectlConfig.Contexts {
		for _, v2 := range toBeAppend.Contexts {
			if strings.EqualFold(v1.Name, v2.Name) {
				return fmt.Errorf("a context entry with %s already exists in kubeconfig, merge failed", v1.Name)
			}
		}
	}

	return nil
}

func Merge(kubeConfig KubectlConfig, toBeAppend KubectlConfig, name string, override bool) (*KubectlConfig, error) {

	var err = ValidateOnlyOneContext(toBeAppend, kubeConfig)
	if err != nil {
		return nil, err
	}

	toBeAppend.Clusters[0].Name = name
	toBeAppend.Users[0].Name = name
	toBeAppend.Contexts[0].Name = name
	toBeAppend.Contexts[0].Context.Cluster = name
	toBeAppend.Contexts[0].Context.User = name

	var needOverride = false
	err = ValidateDuplication(toBeAppend, kubeConfig)
	if err != nil {
		if !override {
			return nil, err
		}
		needOverride = true
	}

	if needOverride {
		// remove the existing cluster, user and context
		for i, v := range kubeConfig.Clusters {
			if strings.EqualFold(v.Name, name) {
				kubeConfig.Clusters = append(kubeConfig.Clusters[:i], kubeConfig.Clusters[i+1:]...)
				break
			}
		}

		for i, v := range kubeConfig.Users {
			if strings.EqualFold(v.Name, name) {
				kubeConfig.Users = append(kubeConfig.Users[:i], kubeConfig.Users[i+1:]...)
				break
			}
		}

		for i, v := range kubeConfig.Contexts {
			if strings.EqualFold(v.Name, name) {
				kubeConfig.Contexts = append(kubeConfig.Contexts[:i], kubeConfig.Contexts[i+1:]...)
				break
			}
		}

		fmt.Printf("Cluster, context and user with '%s' name is removed because of override flag\n", name)
	}

	kubeConfig.Clusters = append(kubeConfig.Clusters, toBeAppend.Clusters[0])
	kubeConfig.Users = append(kubeConfig.Users, toBeAppend.Users[0])
	kubeConfig.Contexts = append(kubeConfig.Contexts, toBeAppend.Contexts[0])

	fmt.Printf("Cluster, context and user will be added with '%s' name\n", name)

	return &kubeConfig, nil
}

func getKubeConfigPath(passedValue string) string {

	//--kubeconfig is not passed
	kubeConfigPath := strings.ReplaceAll(passedValue, " ", "")
	if len(kubeConfigPath) == 0 {
		log.Printf("--kubeconfig was not passed. Looking the %s environment variable\n", KUBECONFIG_ENV)

		// case1: env variable exists
		kubeConfigPath = os.Getenv(KUBECONFIG_ENV)
		// case2: fallback to default path
		if len(kubeConfigPath) == 0 {
			log.Printf("%s env variable does not exist. Default %s path will be used\n", KUBECONFIG_ENV, KUBECONFIG_DEFAULT_PATH)
		}
	}

	return kubeConfigPath
}

func main() {
	kubeConfigPtr := flag.String("kubeconfig", "", fmt.Sprintf("path to the kubeconfig file (defaults '%s' or '%s')", KUBECONFIG_ENV_KEY, KUBECONFIG_DEFAULT_PATH))
	filePtr := flag.String("file", "", "path to the yaml file that to be append into kubeconfig")
	overridePtr := flag.Bool("override", false, "Override the existing context, user and cluster with the file name, or the fields in the file will be used")
	flag.Parse()

	var kubeConfigPath = getKubeConfigPath(*kubeConfigPtr)

	kubeConfig, err := ParseKubeConfig(kubeConfigPath)
	if err != nil {
		log.Panic(err)
	}
	if filepath.Ext(*filePtr) == "" {
		log.Panic("the file specified by --file must have a valid extension")
	}

	toBeAppend, err := ParseKubeConfig(*filePtr)
	if err != nil {
		log.Panic(err)
	}

	var fileName = filepath.Base(*filePtr)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	fileName = strings.ToLower(fileName)
	result, err := Merge(*kubeConfig, *toBeAppend, fileName, *overridePtr)
	if err != nil {
		log.Panic(err)
	}

	data, err := yaml.Marshal(result)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(kubeConfigPath, data, 0644)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("%s was modified successfully\n", kubeConfigPath)
}
