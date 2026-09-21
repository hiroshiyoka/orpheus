package collector

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/alerting"
	"github.com/hiroshiyoka/orpheus/internal/detector"
	"github.com/hiroshiyoka/orpheus/internal/storage"
)

type Pinger func(url string, timeout time.Duration) storage.Check
type AlertSender func(botToken, chatID, message string) error

func Start(ctx context.Context, db *sql.DB, projects []storage.Project, ping Pinger, failureThreshold int, botToken, chatID string, sendAlert AlertSender) {
	if sendAlert == nil {
		sendAlert = alerting.SendTelegramAlert
	}
	var wg sync.WaitGroup
	for _, project := range projects {
		if !project.IsActive {
			continue
		}
		wg.Add(1)
		go func(p storage.Project) {
			defer wg.Done()
			interval := p.CheckIntervalSeconds
			if interval <= 0 {
				interval = 60
			}
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()
			store := func() {
				check := ping(p.URL, 10*time.Second)
				check.ProjectID = p.ID
				saved, err := storage.InsertCheck(db, check)
				if err != nil {
					log.Printf("check insert failed: %v", err)
					return
				}
				n := p.Threshold(failureThreshold)
				if !saved.IsUp {
					incident, err := detector.CheckDowntime(db, p.ID, n)
					if err != nil {
						log.Printf("downtime check failed: %v", err)
					} else if incident != nil {
						message := fmt.Sprintf("[%s] Down\nURL: %s", p.Name, p.URL)
						if err := sendAlert(botToken, chatID, message); err != nil {
							log.Printf("telegram alert failed: %v", err)
						}
					}
				} else {
					incident, err := detector.ResolveDowntime(db, p.ID, true)
					if err != nil {
						log.Printf("resolve check failed: %v", err)
					} else if incident != nil {
						message := fmt.Sprintf("[%s] Recovered\nURL: %s", p.Name, p.URL)
						if err := sendAlert(botToken, chatID, message); err != nil {
							log.Printf("telegram alert failed: %v", err)
						}
					}
				}
			}
			store()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					store()
				}
			}
		}(project)
	}
	wg.Wait()
}
