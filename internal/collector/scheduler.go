package collector

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/detector"
	"github.com/hiroshiyoka/orpheus/internal/storage"
)

type Pinger func(url string, timeout time.Duration) storage.Check

func Start(ctx context.Context, db *sql.DB, projects []storage.Project, ping Pinger, failureThreshold int) {
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
					if _, err := detector.CheckDowntime(db, p.ID, n); err != nil {
						log.Printf("downtime check failed: %v", err)
					}
				} else {
					if _, err := detector.ResolveDowntime(db, p.ID, true); err != nil {
						log.Printf("resolve check failed: %v", err)
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
