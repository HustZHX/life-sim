package service

import (
	"log"
	"time"

	"life-sim/backend/store"
)

// tickJobProgress 在 AI 长耗时调用期间平滑更新进度条，避免前端长期停在 30%。
func tickJobProgress(st *store.Store, jobID string, from, to int, done <-chan struct{}) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	p := from
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if p >= to {
				continue
			}
			p += 4
			if p > to {
				p = to
			}
			if err := st.BumpJobProgressIfRunning(jobID, p); err != nil {
				log.Printf("更新任务进度失败 job=%s: %v", jobID, err)
			}
		}
	}
}
