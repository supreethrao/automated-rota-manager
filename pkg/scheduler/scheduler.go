package scheduler

import (
	"github.com/robfig/cron"
	"time"
)

type Schedule struct {
	cronExpression      string
	execution func()
	cron *cron.Cron
}

func (s *Schedule) Schedule() error {
	err := s.cron.AddFunc(s.cronExpression, s.execution)
	if err != nil {
		return err
	}
	s.cron.Start()
	return nil
}

func (s *Schedule) NextRun() time.Time {
	return s.cron.Entries()[0].Next
}

func NewSchedule(cronExpression string, toExecute func()) *Schedule {
	return &Schedule{
		cronExpression: cronExpression,
		execution:      toExecute,
		cron :          cron.New(),
	}
}
