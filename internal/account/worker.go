package account

import (
	"log"
)

// Notification represents a notification job
type Notification struct {
	AccountID string
	Message   string
}

// NotificationWorkerPool manages async notification tasks
type NotificationWorkerPool struct {
	JobQueue chan Notification
}

// NewNotificationWorkerPool creates a new worker pool
func NewNotificationWorkerPool(bufferSize int) *NotificationWorkerPool {
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

// Enqueue adds a notification job to the queue
func (p *NotificationWorkerPool) Enqueue(notification Notification) {
	select {
	case p.JobQueue <- notification:
	default:
		log.Printf("Warning: notification queue full, dropping notification for %s", notification.AccountID)
	}
}
