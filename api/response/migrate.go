package response

import "time"

// MigrateStatus 迁移任务状态
type MigrateStatus struct {
	RunId       string    `json:"run_id"`
	Status      string    `json:"status"` // running / success / failed / canceled
	RowsRead    int64     `json:"rows_read"`
	RowsWritten int64     `json:"rows_written"`
	Batches     int64     `json:"batches"`
	Message     string    `json:"message"`
	StartedAt   time.Time `json:"started_at"`
	FinishedAt  time.Time `json:"finished_at"`
}
