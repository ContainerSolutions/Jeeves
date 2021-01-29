package scheduler

import (
	jclient "github.com/ContainerSolutions/jeeves/pkg/client"
	"github.com/ContainerSolutions/jeeves/pkg/config"
	jgithub "github.com/ContainerSolutions/jeeves/pkg/github"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

// AddSchedule Runs scheduled tasks for Paul
func AddSchedule(c *cron.Cron) {
	cfg := config.JeevesConfig{}
	cfgErr := cfg.GetConfig()
	if handleErr(cfgErr) {
		return
	}
	_, _ = c.AddFunc("*/30 * * * *", func() {
		gClient, err := jclient.GetInstallationClient(cfg.GithubInstallationID)
		if handleErr(err) {
			return
		}
		jgithub.CheckPullRequests(gClient)
	})
}

func handleErr(err error) bool {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("scheduler error occurred")
		return true
	}
	return false
}
