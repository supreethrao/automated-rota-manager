package helper

import (
	"encoding/binary"
	"gopkg.in/yaml.v2"
	"time"
)

type testTeamMembersYaml struct {
	Members []string `yaml:"members"`
}

var TestTeamMembers = []string{"person1", "person2", "third person"}
var TestTeamMembersListYaml = func() []byte {
	if yml, err := yaml.Marshal(testTeamMembersYaml{TestTeamMembers}); err == nil {
		return yml
	} else {
		panic(err)
	}
}()

func Uint16ToBytes(intVal uint16) []byte {
	byteVal := make([]byte, 2)
	binary.BigEndian.PutUint16(byteVal, intVal)
	return byteVal
}

func Today() string {
	return time.Now().Format("02-01-2006")
}

func Yesterday() string {
	return time.Now().AddDate(0, 0, -1).Format("02-01-2006")
}

func DayBeforeYesterday() string {
	return time.Now().AddDate(0, 0, -2).Format("02-01-2006")
}

func DaysBeforeToday(num int) string {
	return time.Now().AddDate(0, 0, -num).Format("02-01-2006")
}
