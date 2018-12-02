package rota

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type team struct {
	Members []string `yaml:"members"`
}

var members = func() []string {
	val, err := ioutil.ReadFile("/tmp/team.yml")
	members := team{}

	err = yaml.Unmarshal(val, &members)

	if err != nil {
		log.Fatalf("Unable to read the team members file: %v", err)
	}

	return members.Members
}()

func List() []string {
	return members
}
