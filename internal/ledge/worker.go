package ledger

import (
	"log"
)

type Notification struct {
	AccountID string
	Message   string
}

// NotificationWorkerPool manages async tasks
type NotificationWorkerPool struct {
	JobQueue chan Notification
}

func NewWorkerPool(bufferSize int) *NotificationWorkerPool {
	return &NotificationWorkerPool{
		JobQueue: make(chan Notification, bufferSize),
	}
}

// Start spawns N worker goroutines
func (p *NotificationWorkerPool) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for job := range p.JobQueue {
				// Simulating external API call (Email/SMS)
				log.Printf("Worker %d: Sending notification to %s: %s", id, job.AccountID, job.Message)
			}
		}(i)
	}
}
