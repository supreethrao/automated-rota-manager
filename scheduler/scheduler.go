package scheduler

import "github.com/robfig/cron"

type Schedule interface {
	Schedule()
}

type schedule struct {
	cron      string
	execution func()
}

func (s *schedule) Schedule() {
	crony := cron.New()
	crony.AddFunc(s.cron, s.execution)
	crony.Start()
}

func NewSchedule(cronExpression string, todo func()) Schedule {
	return &schedule{
		cron:      cronExpression,
		execution: todo,
	}
}
