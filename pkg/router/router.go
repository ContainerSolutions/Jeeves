package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"

	"github.com/ContainerSolutions/jeeves/pkg/helpers"
	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//GetRouter .
func GetRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", SlackEventHandler)
	return r
}

func SlackEventHandler(w http.ResponseWriter, r *http.Request) {
	verificationToken := helpers.GetEnv("SLACK_VERIFICATION_TOKEN", "")
	authToken := helpers.GetEnv("SLACK_AUTHENTICATION_TOKEN", "")

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	body := buf.String()
	eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: verificationToken}))
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r *slackevents.ChallengeResponse
		err := json.Unmarshal([]byte(body), &r)
		if handleError(err) {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "text")
		w.Write([]byte(r.Challenge))
	}
	var api = slack.New(authToken)
	if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			_, _, _ = api.PostMessage(ev.Channel, slack.MsgOptionText("Anonymizing Repo", false))
			url, candidateId, err := getLinkAndId(ev.Text)
			err = createAnonymizastionJob(url, candidateId)
			if handleError(err) {
				api.PostMessage(ev.Channel, slack.MsgOptionText("There was an Error Anonymizing Repo Please Contact Brendan", false))
				w.WriteHeader(http.StatusOK)
				return
			}
			_, _, _ = api.PostMessage(ev.Channel, slack.MsgOptionText("API Anonymized", false))

		}
	}
}

func handleError(err error) bool {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("webhook error occurred")
		return true
	}
	return false
}

func getLinkAndId(Message string) (string, string, error) {
	log.Printf("%v", Message)
	res := strings.Replace(Message, "<https://gitlab.com/", "", -1)
	res = strings.Replace(res, ">\u00a0", " ", -1)
	args := strings.Split(res, " ")
	log.Printf("%v", args)
	if len(args) != 3 {
		return "", "", fmt.Errorf("Wrong Format")
	}
	return args[1], args[2], nil
}

func createAnonymizastionJob(link string, candidateId string) error {
	clientset, _ := GetClientSet()
	jobsClient := clientset.BatchV1().Jobs("jeeves")
	commonMeta := metav1.ObjectMeta{
		Name:      strings.ToLower(candidateId),
		Namespace: "jeeves",
	}
	var secretMode int32 = 0600

	job := &batchv1.Job{
		ObjectMeta: commonMeta,
		Spec: batchv1.JobSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					RestartPolicy: "Never",
					Volumes: []apiv1.Volume{
						apiv1.Volume{
							Name: "sshkey",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: "sshkey",
									Items: []apiv1.KeyToPath{
										apiv1.KeyToPath{Key: "id_rsa", Path: "id_rsa"},
									},
									DefaultMode: &secretMode,
								},
							},
						},
						apiv1.Volume{
							Name: "credentials",
							VolumeSource: apiv1.VolumeSource{
								Secret: &apiv1.SecretVolumeSource{
									SecretName: "credentials",
									Items: []apiv1.KeyToPath{
										apiv1.KeyToPath{Key: "credentials.json", Path: "credentials.json"},
									},
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "anonymizer",
							Image: "containersol/anonymizer:latest",
							Args: []string{
								link,
								candidateId,
							},
							Env: []apiv1.EnvVar{
								apiv1.EnvVar{
									Name:  "GOOGLE_APPLICATION_CREDENTIALS",
									Value: "/infra/.user/credentials/credentials.json",
								},
								apiv1.EnvVar{
									Name:  "CS_REVIEWER_KEY",
									Value: "/infra/.user/.ssh/id_rsa",
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								apiv1.VolumeMount{
									Name:      "sshkey",
									MountPath: "/infra/.user/.ssh/",
									ReadOnly:  true,
								},
								apiv1.VolumeMount{
									Name:      "credentials",
									MountPath: "/infra/.user/credentials/",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := jobsClient.Create(context.Background(), job, metav1.CreateOptions{})
	return err
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
