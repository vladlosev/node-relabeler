package kube

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EnvRestore struct {
	Restore func()
}

func restoreEnvVar(name, value string, exists bool) EnvRestore {
	return EnvRestore{Restore: func() {
		if exists {
			os.Setenv(name, value)
		} else {
			os.Unsetenv(name)
		}
	}}
}

func SetEnvVar(name, value string) EnvRestore {
	oldValue, exists := os.LookupEnv(name)
	os.Setenv(name, value)
	return restoreEnvVar(name, oldValue, exists)
}

func UnsetEnvVar(name string) EnvRestore {
	oldValue, exists := os.LookupEnv(name)
	os.Unsetenv(name)
	return restoreEnvVar(name, oldValue, exists)
}

const sampleConfig = `
apiVersion: v1
kind: Config
current-context: dev-frontend
clusters:
- cluster:
    server: https://remote-server
  name: development
users:
- name: developer
contexts:
- context:
    cluster: development
    user: developer
  name: dev-frontend
`

func TestConfigFromKubeConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	fileName := filepath.Join(dir, "kubeconfig")
	err = ioutil.WriteFile(fileName, []byte(sampleConfig), 0666)
	require.NoError(t, err)
	defer SetEnvVar("KUBECONFIG", fileName).Restore()

	config, err := getConfig()
	assert.NoError(t, err)
	assert.Equal(t, config.Host, "https://remote-server")
}

func TestConfigFromHome(t *testing.T) {
	homeDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(homeDir)
	dir := filepath.Join(homeDir, ".kube")
	err = os.MkdirAll(dir, 0777)
	require.NoError(t, err)
	fileName := filepath.Join(dir, "config")
	err = ioutil.WriteFile(fileName, []byte(sampleConfig), 0666)
	require.NoError(t, err)
	defer UnsetEnvVar("KUBECONFIG").Restore()
	defer SetEnvVar("HOME", homeDir).Restore()

	config, err := getConfig()
	assert.NoError(t, err)
	assert.Equal(t, config.Host, "https://remote-server")
}

func TestConfigInCluster(t *testing.T) {
	homeDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(homeDir)
	dir := filepath.Join(homeDir, ".kube")
	err = os.MkdirAll(dir, 0777)
	require.NoError(t, err)
	defer UnsetEnvVar("KUBECONFIG").Restore()
	defer SetEnvVar("HOME", homeDir).Restore()
	defer SetEnvVar("KUBERNETES_SERVICE_HOST", "master").Restore()
	defer SetEnvVar("KUBERNETES_SERVICE_PORT", "443").Restore()

	_, err = getConfig()
	if err != nil {
		// It's not possible to check for the in cluster config being created
		// correctly outside of a cluster. We just fish for a right error message.
		assert.Regexp(
			t,
			"open .*/serviceaccount/token: no such file or directory",
			err.Error(),
		)
	}
}
