package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ContainerSolutions/jeeves/pkg/helpers"
	"github.com/slack-go/slack"
	"k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

const (
	secretKeyFile  = "jeeves-secret-key"
	privateKeyFile = "jeeves-private-key"
)

type JeevesConfig struct {
	JobNamespace        string               `json:",omitempty"`
	SlackApi            *slack.Client        `json:",omitempty"`
	SlackSigningToken   string               `json:",omitempty"`
	K8sClientSet        kubernetes.Interface `json:",omitempty"`
	GithubSecretKey     string               `json:",omitempty"`
	GithubPrivateKey    string               `json:",omitempty"`
	GithubApplicationID string               `json:",omitempty"`
}

func (cfg *JeevesConfig) GetConfig() error {
	// Set the namespace for Kubernetes Jobs
	cfg.JobNamespace = helpers.GetEnv("NAMESPACE", "")

	// Set Kubernetes Client
	testKubernetes := helpers.GetEnv("TEST_KUBERNETES", "")
	if testKubernetes == "" {
		clientset, _ := GetClientSet()
		cfg.K8sClientSet = clientset
	} else {
		cfg.K8sClientSet = fake.NewSimpleClientset()
	}

	// Set Slack Signing Token
	cfg.SlackSigningToken = helpers.GetEnv("SLACK_SIGNING_SECRET", "")

	// Set Slack Client
	authToken := helpers.GetEnv("SLACK_AUTHENTICATION_TOKEN", "")
	testSlackAPIURL := helpers.GetEnv("TEST_SLACK_URL", "")
	if testSlackAPIURL == "" {
		cfg.SlackApi = slack.New(authToken)
	} else {
		cfg.SlackApi = slack.New(
			authToken,
			slack.OptionAPIURL("http://"+testSlackAPIURL+"/"),
		)
	}

	// Set Github Secret Key
	keyPath, pathErr := getSecretPath()
	if pathErr != nil {
		return pathErr
	}
	secretKeyBytes, readErr := ioutil.ReadFile(path.Join(keyPath, secretKeyFile))
	if readErr != nil {
		return fmt.Errorf(
			"unable to read GitHub symmetrical secret: %s, error: %s",
			keyPath+secretKeyFile,
			readErr,
		)
	}
	secretKeyBytes = getFirstLine(secretKeyBytes)
	cfg.GithubSecretKey = string(secretKeyBytes)

	// Set Github RSA Private Key
	privateKeyPath := path.Join(keyPath, privateKeyFile)
	keyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf(
			"unable to read private key path: %s, error: %s",
			privateKeyPath,
			err,
		)
	}
	cfg.GithubPrivateKey = string(keyBytes)

	// Set Github Application ID
	if val, ok := os.LookupEnv("APPLICATION_ID"); ok && len(val) > 0 {
		cfg.GithubApplicationID = val
	} else {
		return fmt.Errorf("APPLICATION_ID must be given")
	}

	return nil
}

/*
   GetClientSet Generates a clientset from the Kubeconfig
*/
func GetClientSet() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func getSecretPath() (string, error) {
	secretPath := os.Getenv("SECRET_PATH")

	if len(secretPath) == 0 {
		return "", fmt.Errorf("SECRET_PATH env-var not set")
	}

	return secretPath, nil
}

func getFirstLine(secret []byte) []byte {
	stringSecret := string(secret)
	if newLine := strings.Index(stringSecret, "\n"); newLine != -1 {
		secret = secret[:newLine]
	}
	return secret
}
