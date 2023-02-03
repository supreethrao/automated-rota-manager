package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ARMConfig struct {
	CronSchedule string	`yaml:"cron_schedule"`
	TeamName string	`yaml:"team_name"`
	IngressURL string	`yaml:"ingress_url"`
	SlackUserName string `yaml:"slack_user_name"`
	SlackChannel string	`yaml:"slack_channel"`
}

func New(filePath string) (*ARMConfig, error){
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg ARMConfig

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
