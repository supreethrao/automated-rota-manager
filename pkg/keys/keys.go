package keys

import (
	"time"
)

type Keys struct {
	rootPrefix string
}

func (key *Keys) TeamKey() string {
	return key.rootPrefix + "::team_members"
}

func (key *Keys) AccruedDaysCounterKey(memberName string) string {
	return key.rootPrefix + "::member::" + memberName
}

func (key *Keys) PersonPickedOnDayKey(whichDay time.Time) string {
	formattedDay := whichDay.Format("02-01-2006")
	return key.rootPrefix + "::" + formattedDay
}

func (key *Keys) LatestDayPickedKey(memberName string) string {
	return key.rootPrefix + "::latest-day::" + memberName
}

func (key *Keys) LatestCronRunKey() string {
	return key.rootPrefix + "::latest-cron"
}

func (key *Keys) OutOfOfficeKey(memberName string) (string, string) {
	keyBase := key.rootPrefix + "::out_of_office::" + memberName
	return keyBase + "::from_date", keyBase + "::to_date"
}

func NewKey(rootPrefix string) Keys {
	return Keys{rootPrefix}
}
