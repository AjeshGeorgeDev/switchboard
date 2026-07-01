package scheduler

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/switchboard/switchboard/internal/config"
	"github.com/switchboard/switchboard/internal/jobs"
)

func Start(cfg config.Config, client *asynq.Client) *cron.Cron {
	c := cron.New()
	if cfg.CVEPullEnabled {
		_, err := c.AddFunc(cfg.CVEPullCron, func() {
			if _, err := client.Enqueue(asynq.NewTask(jobs.TypeCVEPull, nil)); err != nil {
				log.Printf("enqueue cve pull: %v", err)
			}
		})
		if err != nil {
			log.Printf("cron cve pull: %v", err)
		} else {
			log.Printf("CVE pull cron enabled: %s", cfg.CVEPullCron)
		}
	} else {
		log.Printf("CVE pull cron disabled (set CVE_PULL_ENABLED=true to enable)")
	}

	_, err := c.AddFunc("0 3 * * *", func() {
		if _, err := client.Enqueue(asynq.NewTask(jobs.TypeRetentionCleanup, nil)); err != nil {
			log.Printf("enqueue retention cleanup: %v", err)
		}
	})
	if err != nil {
		log.Printf("cron retention: %v", err)
	}

	c.Start()
	return c
}
