package collector

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/hiroshiyoka/orpheus/internal/storage"
)

type Pinger func(url string, timeout time.Duration) storage.Check

func Start(ctx context.Context, db *sql.DB, projects []storage.Project, ping Pinger) {
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
				if _, err := storage.InsertCheck(db, check); err != nil {
					log.Printf("check insert failed: %v", err)
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
