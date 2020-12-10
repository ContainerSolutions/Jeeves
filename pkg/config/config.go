package config

import (
	"github.com/ContainerSolutions/jeeves/pkg/helpers"
	"github.com/slack-go/slack"
	"k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type JeevesConfig struct {
	Namespace    string               `json:",omitempty"`
	SlackApi     *slack.Client        `json:",omitempty"`
	SigningToken string               `json:",omitempty"`
	ClientSet    kubernetes.Interface `json:",omitempty"`
}

func (cfg *JeevesConfig) GetConfig() {
	authToken := helpers.GetEnv("SLACK_AUTHENTICATION_TOKEN", "")
	testSlackAPIURL := helpers.GetEnv("TEST_SLACK_URL", "")
	testKubernetes := helpers.GetEnv("TEST_KUBERNETES", "")
	cfg.Namespace = helpers.GetEnv("NAMESPACE", "")
	if testSlackAPIURL == "" {
		cfg.SlackApi = slack.New(authToken)
	} else {
		cfg.SlackApi = slack.New(
			authToken,
			slack.OptionAPIURL("http://"+testSlackAPIURL+"/"),
		)
	}
	if testKubernetes == "" {
		clientset, _ := GetClientSet()
		cfg.ClientSet = clientset
	} else {
		cfg.ClientSet = fake.NewSimpleClientset()
	}
	cfg.SigningToken = helpers.GetEnv("SLACK_SIGNING_SECRET", "")
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
