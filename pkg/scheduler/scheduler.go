package scheduler

import (
	"github.com/robfig/cron"
)

type Schedule struct {
	cron      string
	execution func()
}

func (s *Schedule) Schedule() error {
	crony := cron.New()
	err := crony.AddFunc(s.cron, s.execution)
	if err != nil {
		return err
	}
	crony.Start()
	return nil
}

func NewSchedule(cronExpression string, todo func()) *Schedule {
	return &Schedule{
		cron:      cronExpression,
		execution: todo,
	}
}

