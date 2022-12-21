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
const KUBECONFIG_DEFAULT_PATH = "~/kube/config"

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

func validate(kubectlConfig KubectlConfig, name string) error {

	if len(kubectlConfig.Clusters) != 1 {
		return errors.New("Only one cluster can be merged into original kubeconfig")
	}

	if len(kubectlConfig.Users) != 1 {
		return errors.New("Only one user can be merged into original kubeconfig")
	}

	if len(kubectlConfig.Contexts) != 1 {
		return errors.New("Only one context can be merged into original kubeconfig")
	}

	for _, v := range kubectlConfig.Clusters {
		if strings.EqualFold(v.Name, name) {
			return errors.New(fmt.Sprintf("A cluster entry with %s already exists in kubeconfig, merge failed!", name))
		}
	}

	for _, v := range kubectlConfig.Contexts {
		if strings.EqualFold(v.Name, name) {
			return errors.New(fmt.Sprintf("A context entry with %s already exists in kubeconfig, merge failed!", name))
		}
	}

	for _, v := range kubectlConfig.Users {
		if strings.EqualFold(v.Name, name) {
			return errors.New(fmt.Sprintf("A user entry with %s already exists in kubeconfig, merge failed!", name))
		}
	}

	return nil
}

func Merge(kubeConfig KubectlConfig, toBeAppend KubectlConfig, name, toBeAppendFileName string) (*KubectlConfig, error) {

	var newName = toBeAppendFileName
	if len(name) > 0 {
		newName = name
	}

	var err = validate(toBeAppend, toBeAppendFileName)
	if err != nil {
		return nil, err
	}

	toBeAppend.Clusters[0].Name = newName
	toBeAppend.Users[0].Name = newName
	toBeAppend.Contexts[0].Name = newName
	toBeAppend.Contexts[0].Context.Cluster = newName
	toBeAppend.Contexts[0].Context.User = newName
	kubeConfig.Clusters = append(kubeConfig.Clusters, toBeAppend.Clusters[0])
	kubeConfig.Users = append(kubeConfig.Users, toBeAppend.Users[0])
	kubeConfig.Contexts = append(kubeConfig.Contexts, toBeAppend.Contexts[0])

	fmt.Printf("Cluster, context and user will be added with '%s' name\n", newName)

	return &kubeConfig, nil
}

func getKubeConfigPath(passedValue string) string {

	//--kubeconfig is not passed
	kubeConfigPath := strings.ReplaceAll(passedValue, " ", "")
	if len(kubeConfigPath) == 0 {
		log.Printf("--kubeconfig was not passed. Looking the %s environment variable\n", KUBECONFIG_ENV)

		// case1: env variable exists
		kubeConfigPath = os.Getenv(KUBECONFIG_ENV)
		if len(kubeConfigPath) == 0 {
			log.Fatalf("%s exists with no value!", KUBECONFIG_ENV)
		} else {
			// case2: fallback to default path
			log.Printf("%s env variable does not exist. Default %s path will be used\n", KUBECONFIG_ENV, KUBECONFIG_DEFAULT_PATH)
		}
	}

	return kubeConfigPath
}

func main() {
	kubeConfigPtr := flag.String("kubeconfig", "", fmt.Sprintf("path to the kubeconfig file (defaults '%s' or '%s')", KUBECONFIG_ENV_KEY, KUBECONFIG_DEFAULT_PATH))
	filePtr := flag.String("file", "", "path to the yaml file that to be append into kubeconfig")
	namePtr := flag.String("name", "", fmt.Sprintf("Replaces the name of context, user and cluster (default file name of --file argument)"))
	flag.Parse()

	var kubeConfigPath = getKubeConfigPath(*kubeConfigPtr)

	kubeConfig, err := ParseKubeConfig(kubeConfigPath)
	if err != nil {
		log.Panic(err)
	}

	toBeAppend, err := ParseKubeConfig(*filePtr)
	if err != nil {
		log.Panic(err)
	}

	var fileName = filepath.Base(*filePtr)
	result, err := Merge(*kubeConfig, *toBeAppend, *namePtr, fileName[:len(fileName)-len(filepath.Ext(fileName))])
	if err != nil {
		log.Panic(err)
	}

	data, err := yaml.Marshal(result)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(kubeConfigPath, data, 0644)
	fmt.Printf("%s was modified successfully\n", kubeConfigPath)
}
