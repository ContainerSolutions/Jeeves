package commands

import (
	"github.com/ContainerSolutions/jeeves/pkg/config"
	"github.com/stretchr/testify/assert"
	"strings"

	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

const jeevesURL = "https://jeeves.test.example/anonymize"

func TestAnonymizeHandler(t *testing.T) {
	mux, serverUrl, teardown := GetMockClient()
	defer teardown()

	os.Setenv("TEST_SLACK_URL", serverUrl)
	os.Setenv("TEST_KUBERNETES", "True")
	t.Run("Test Anonymize Handler Successful", func(t *testing.T) {
		cfg := config.JeevesConfig{}
		mux.HandleFunc(
			"/chat.postMessage",
			func(w http.ResponseWriter, r *http.Request) {},
		)
		cfg.GetConfig()
		body := url.Values{
			"command":         []string{"/anonymize"},
			"team_domain":     []string{"team"},
			"enterprise_id":   []string{"E0001"},
			"enterprise_name": []string{"Globular%20Construct%20Inc"},
			"channel_id":      []string{"C1234ABCD"},
			"text":            []string{"https://gitlab.com/random-user/api-excercise Test00001"},
			"team_id":         []string{"T1234ABCD"},
			"user_id":         []string{"U1234ABCD"},
			"user_name":       []string{"username"},
			"response_url":    []string{"https://hooks.slack.com/commands/XXXXXXXX/00000000000/YYYYYYYYYYYYYY"},
			"token":           []string{"valid"},
			"channel_name":    []string{"channel"},
			"trigger_id":      []string{"0000000000.1111111111.222222222222aaaaaaaaaaaaaa"},
			"api_app_id":      []string{"A123456"},
		}
		req, err := http.NewRequest(
			http.MethodPost,
			jeevesURL,
			strings.NewReader(body.Encode()),
		)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		assert.Equal(t, nil, err)
		err = AnonymizeHandler(&cfg, req)
		assert.Equal(t, nil, err)
	})
}

func TestGetLinkAndId(t *testing.T) {
	t.Run("Test Parse Args Succesful - Gitlab", func(t *testing.T) {
		message := "https://gitlab.com/random-user/api-excercise Test00001"
		repoType, repo, candidateId, err := parseArgs(message)
		assert.Equal(t, nil, err)
		assert.Equal(t, "gitlab", repoType)
		assert.Equal(t, "random-user/api-excercise", repo)
		assert.Equal(t, "Test00001", candidateId)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Parse Args Succesful - Github", func(t *testing.T) {
		message := "<https://github.com/random-user/api-excercise>\u00a0Test00001"
		repoType, repo, candidateId, err := parseArgs(message)
		assert.Equal(t, nil, err)
		assert.Equal(t, "github", repoType)
		assert.Equal(t, "random-user/api-excercise", repo)
		assert.Equal(t, "Test00001", candidateId)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Parse Args Succesful - Double Spaces", func(t *testing.T) {
		message := "<https://github.com/random-user/api-excercise>\u00a0\u00a0Test00001"
		repoType, repo, candidateId, err := parseArgs(message)
		assert.Equal(t, nil, err)
		assert.Equal(t, "github", repoType)
		assert.Equal(t, "random-user/api-excercise", repo)
		assert.Equal(t, "Test00001", candidateId)
		assert.Equal(t, nil, err)
	})
	t.Run("Test Parse Args Succesful - Uppercase", func(t *testing.T) {
		message := "<https://github.com/random-User/Api-Excercise>\u00a0\u00a0Test00001"
		repoType, repo, candidateId, err := parseArgs(message)
		assert.Equal(t, nil, err)
		assert.Equal(t, "github", repoType)
		assert.Equal(t, "random-user/api-excercise", repo)
		assert.Equal(t, "Test00001", candidateId)
		assert.Equal(t, nil, err)
	})
}

//GetMockClient Returns a Mock Client in order to mock out calls to Githubs API
func GetMockClient() (
	mux *http.ServeMux,
	serverURL string,
	teardown func(),
) {
	mux = http.NewServeMux()
	server := httptest.NewServer(nil)

	return mux, server.URL, server.Close
}
