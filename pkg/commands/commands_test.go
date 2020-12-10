package commands

import (
	"bytes"
	"github.com/ContainerSolutions/jeeves/pkg/config"
	"github.com/stretchr/testify/assert"

	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const jeevesURL = "https://jeeves.test.example/anonymize"

func TestClientGetCat(t *testing.T) {
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
		body := []byte(`
        {
            "token": "gIkuvaNzQIHg97ATvDxqgjtO",
            "team_id": "T0001",
            "team_domain": "example",
            "enterprise_id": "E0001",
            "enterprise_name": "TEST",
            "channel_id": "C2147483705",
            "channel_name": "test",
            "user_id": "U2147483697",
            "user_name": "Steve",
            "command": "/weather",
            "text": "https://gitlab.com/user/api-excercies TEST0001",
            "response_url": "https://hooks.slack.com/commands/1234/5678",
            "trigger_id": "13345224609.738474920.8088930838d88f008e0",
            "api_app_id": "A123456"
        }
        `)
		request, err := http.NewRequest(
			http.MethodPost,
			jeevesURL,
			bytes.NewReader(body),
		)
		assert.Equal(t, nil, err)
		err = AnonymizeHandler(&cfg, request)
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
