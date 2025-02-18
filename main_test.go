package main

import (
	"log"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func getKubeConfig() *KubectlConfig {
	const KUBECONFIG_FILENAME = "kubeconfig.yaml"
	return getKubeConfigFromPath(KUBECONFIG_FILENAME)
}

func getKubeConfigFromPath(filename string) *KubectlConfig {
	const TEST_DATA_PATH = "test/data"
	kubeConfig, err := ParseKubeConfig(filepath.Join(TEST_DATA_PATH, filename))
	if err != nil {
		log.Panic(err)
	}

	return kubeConfig
}

// cluster name conflict
func Test_Merge_ClusterNameConflict(t *testing.T) {
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath("valid-default-cluster.yaml")

	toBeAddedKubeConfig.Clusters[0].Name = kubeConfig.Clusters[0].Name

	_, err := Merge(*kubeConfig, *toBeAddedKubeConfig, toBeAddedKubeConfig.Clusters[0].Name, false)
	if err == nil {
		t.Fatal()
	}
}

// context name conflict
func Test_Merge_ContextNameConflict(t *testing.T) {
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath("valid-default-cluster.yaml")

	toBeAddedKubeConfig.Contexts[0].Name = kubeConfig.Contexts[0].Name

	_, err := Merge(*kubeConfig, *toBeAddedKubeConfig, toBeAddedKubeConfig.Contexts[0].Name, false)
	if err == nil {
		t.Fatal()
	}
}

// user name conflict
func Test_Merge_UserNameConflict(t *testing.T) {
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath("valid-default-cluster.yaml")

	toBeAddedKubeConfig.Users[0].Name = kubeConfig.Users[0].Name

	_, err := Merge(*kubeConfig, *toBeAddedKubeConfig, toBeAddedKubeConfig.Users[0].Name, false)
	if err == nil {
		t.Fatal()
	}
}

// multiple cluster
func Test_Merge_MultipleCluster(t *testing.T) {
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath("valid-default-cluster.yaml")

	var newObj KubectlClusterWithName
	newObj.Name = "new-cluster"
	toBeAddedKubeConfig.Clusters = append(toBeAddedKubeConfig.Clusters, &newObj)

	_, err := Merge(*kubeConfig, *toBeAddedKubeConfig, toBeAddedKubeConfig.Clusters[0].Name, false)
	if err == nil {
		t.Fatal()
	}
}

// multiple context
func Test_Merge_MultipleContext(t *testing.T) {
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath("valid-default-cluster.yaml")

	var newObj KubectlContextWithName
	newObj.Name = "new-context"
	toBeAddedKubeConfig.Contexts = append(toBeAddedKubeConfig.Contexts, &newObj)

	_, err := Merge(*kubeConfig, *toBeAddedKubeConfig, toBeAddedKubeConfig.Clusters[0].Name, false)
	if err == nil {
		t.Fatal()
	}
}

// multiple user
func Test_Merge_MultipleUser(t *testing.T) {
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath("valid-default-cluster.yaml")

	var newObj KubectlUserWithName
	newObj.Name = "new-user"
	toBeAddedKubeConfig.Users = append(toBeAddedKubeConfig.Users, &newObj)

	_, err := Merge(*kubeConfig, *toBeAddedKubeConfig, toBeAddedKubeConfig.Clusters[0].Name, false)
	if err == nil {
		t.Fatal()
	}
}

func Test_Merge_ExplicitlySetName(t *testing.T) {
	const fileName = "valid-default-cluster"
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath(fileName + ".yaml")
	var countBeforeMerge = len(kubeConfig.Clusters)

	var explicitName = uuid.New().String()
	toBeAddedKubeConfig.Clusters[0].Name = explicitName
	toBeAddedKubeConfig.Contexts[0].Name = explicitName
	toBeAddedKubeConfig.Users[0].Name = explicitName

	result, err := Merge(*kubeConfig, *toBeAddedKubeConfig, explicitName, false)
	assert.NoError(t, err)

	assert.Equal(t, len(result.Contexts), countBeforeMerge+1)
	assert.Equal(t, len(result.Clusters), countBeforeMerge+1)
	assert.Equal(t, len(result.Users), countBeforeMerge+1)

	assert.EqualValues(t, result.Contexts[countBeforeMerge].Name, explicitName)
	assert.EqualValues(t, result.Clusters[countBeforeMerge].Name, explicitName)
	assert.EqualValues(t, result.Users[countBeforeMerge].Name, explicitName)
}

func Test_Merge_ImplicitlySetName(t *testing.T) {
	const fileName = "valid-default-cluster"
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath(fileName + ".yaml")
	var countBeforeMerge = len(kubeConfig.Clusters)

	var explicitName = uuid.New().String()
	toBeAddedKubeConfig.Clusters[0].Name = explicitName
	toBeAddedKubeConfig.Contexts[0].Name = explicitName
	toBeAddedKubeConfig.Users[0].Name = explicitName

	result, err := Merge(*kubeConfig, *toBeAddedKubeConfig, fileName, false)
	assert.NoError(t, err)

	assert.Equal(t, len(result.Contexts), countBeforeMerge+1)
	assert.Equal(t, len(result.Clusters), countBeforeMerge+1)
	assert.Equal(t, len(result.Users), countBeforeMerge+1)

	assert.EqualValues(t, result.Contexts[countBeforeMerge].Name, fileName)
	assert.EqualValues(t, result.Clusters[countBeforeMerge].Name, fileName)
	assert.EqualValues(t, result.Users[countBeforeMerge].Name, fileName)
}

func Test_Merge_DuplicatedContextClusterUser(t *testing.T) {
	const fileName = "cluster-1"
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath(fileName + ".yaml")

	result, err := Merge(*kubeConfig, *toBeAddedKubeConfig, fileName, false)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_Merge_DuplicatedContextClusterUser_WithOverride(t *testing.T) {
	const fileName = "cluster-1"
	var kubeConfig = getKubeConfig()
	var toBeAddedKubeConfig = getKubeConfigFromPath(fileName + ".yaml")

	result, err := Merge(*kubeConfig, *toBeAddedKubeConfig, fileName, true)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func Test_ParseKubeConfig(t *testing.T) {
	const fileName = "kubeconfig.yaml"
	kubeConfig, err := ParseKubeConfig(filepath.Join("test/data", fileName))
	assert.NoError(t, err)
	assert.NotNil(t, kubeConfig)
}

func Test_ParseKubeConfig_NotExistingFile(t *testing.T) {
	const fileName = "not-existing-file.yaml"
	kubeConfig, err := ParseKubeConfig(filepath.Join("test/data", fileName))
	assert.Error(t, err)
	assert.Nil(t, kubeConfig)
}

func Test_ValidateOnlyOneContext(t *testing.T) {
	var kubeConfig = getKubeConfig()
	var kc1 = getKubeConfigFromPath("valid-default-cluster.yaml")
	var kc2 = getKubeConfigFromPath("valid-non-default-cluster.yaml")

	//merge them
	kc1.Clusters = append(kc1.Clusters, kc2.Clusters...)
	kc1.Users = append(kc1.Users, kc2.Users...)
	kc1.Contexts = append(kc1.Contexts, kc2.Contexts...)

	err := ValidateOnlyOneContext(*kc1, *kubeConfig)
	assert.Error(t, err)
}

func Test_ValidateDuplication(t *testing.T) {
	var kc1 = getKubeConfig()
	var kc2 = getKubeConfigFromPath("cluster-1.yaml")

	err := ValidateDuplication(*kc1, *kc2)
	assert.Error(t, err)
}
