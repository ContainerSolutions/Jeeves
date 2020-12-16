package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJeevesConfig(t *testing.T) {
	tmpDir := os.TempDir()
	reader := rand.Reader
	bitSize := 4096
	key, err := rsa.GenerateKey(reader, bitSize)
	assert.Equal(t, err, nil)
	privateWant := "private"
	ioutil.WriteFile(path.Join(tmpDir, "jeeves-secret-key"), []byte(privateWant), 0600)
	pemSecretfile, err := os.Create(path.Join(tmpDir, "jeeves-private-key"))
	assert.Equal(t, err, nil)
	defer pemSecretfile.Close()

	err = pem.Encode(
		pemSecretfile,
		&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	assert.Equal(t, nil, err)

	defer os.RemoveAll(path.Join(tmpDir, "paul-private-key"))
	defer os.RemoveAll(path.Join(tmpDir, "paul-secret-key"))

	cfg := JeevesConfig{}
	t.Run("Test Set Jeeves Config With No Settings Error", func(t *testing.T) {
		err := cfg.GetConfig()
		assert.NotEqual(t, nil, err)
	})
	os.Setenv("SECRET_PATH", tmpDir)
	os.Setenv("APPLICATION_ID", "1234")
	os.Setenv("NAMESPACE", "test")
	err = cfg.GetConfig()
	assert.Equal(t, nil, err)
	t.Run("Check Value - JobNamspace", func(t *testing.T) {
		assert.Equal(t, "test", cfg.JobNamespace)
	})
	t.Run("Check Value - GithubApplicationID", func(t *testing.T) {
		assert.Equal(t, "1234", cfg.GithubApplicationID)
	})
	t.Run("Check Value - GithubPrivateKey", func(t *testing.T) {
		assert.NotEqual(t, nil, cfg.GithubPrivateKey)
	})
	t.Run("Check Value - GithubSecretKey", func(t *testing.T) {
		assert.Equal(t, privateWant, cfg.GithubSecretKey)
	})
}
