package model

import "time"

type InstanceStatus string

const (
	InstanceStatusPending     InstanceStatus = "pending"
	InstanceStatusRunning     InstanceStatus = "running"
	InstanceStatusTerminating InstanceStatus = "terminating"
)

type Instance struct {
	InstanceID   string
	UserID       string
	Type         string
	DisplayName  string
	PodName      string
	PodIP        string
	Status       InstanceStatus
	LastActiveAt time.Time
}