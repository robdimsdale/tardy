package tardy

import "time"

type Task struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	DueDate     time.Time `json:"due_date"`
	CompletedAt time.Time `json:"completed_at"`
	Days        int       `json:"days":`
}
