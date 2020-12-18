package kubernetes

import (
	"context"
	"strings"

	"github.com/ContainerSolutions/jeeves/pkg/config"
	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateAnonymizastionJob(
	cfg *config.JeevesConfig,
	repoType, repo, candidateId string,
) error {
	log.Printf("%v", cfg.JobNamespace)
	jobsClient := cfg.K8sClientSet.BatchV1().Jobs(cfg.JobNamespace)
	commonMeta := metav1.ObjectMeta{
		Name:      strings.ToLower(candidateId),
		Namespace: cfg.JobNamespace,
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
										apiv1.KeyToPath{
											Key:  "id_rsa",
											Path: "id_rsa",
										},
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
										apiv1.KeyToPath{
											Key:  "credentials.json",
											Path: "credentials.json",
										},
									},
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "anonymizer",
							Image: "containersol/anonymizer:v1.0.0",
							Args: []string{
								repoType,
								repo,
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
	_, err := jobsClient.Create(
		context.Background(),
		job,
		metav1.CreateOptions{},
	)
	return err
}
