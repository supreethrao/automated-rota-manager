package keys

import (
	"time"
)

type Keys interface {
	TeamKey() string
	SupportDaysCounterKey(memberName string) string
	SupportPersonOnDayKey(supportDay time.Time) string
	LatestDayOnSupportKey(memberName string) string
	OutOfOfficeKey(memberName string) (string, string)
}

type keys struct {
	rootPrefix string
}

func (key *keys) TeamKey() string {
	return key.rootPrefix + "::team_members"
}

func (key *keys) SupportDaysCounterKey(memberName string) string {
	return key.rootPrefix + "::member::" + memberName
}

func (key *keys) SupportPersonOnDayKey(supportDay time.Time) string {
	formattedDay := supportDay.Format("02-01-2006")
	return key.rootPrefix + "::" + formattedDay
}

func (key *keys) LatestDayOnSupportKey(memberName string) string {
	return key.rootPrefix + "::latest-day::" + memberName
}

func (key *keys) OutOfOfficeKey(memberName string) (string, string) {
	keyBase := key.rootPrefix + "::out_of_office::" + memberName
	return keyBase + "::from_date", keyBase + "::to_date"
}

func NewKey(rootPrefix string) Keys {
	return &keys{rootPrefix}
}
