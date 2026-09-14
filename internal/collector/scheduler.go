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
			ticker := time.NewTicker(time.Duration(p.CheckIntervalSeconds) * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					check := ping(p.URL, 10*time.Second)
					check.ProjectID = p.ID
					if _, err := storage.InsertCheck(db, check); err != nil {
						log.Printf("check insert failed: %v", err)
					}
				}
			}
		}(project)
	}
	wg.Wait()
}
