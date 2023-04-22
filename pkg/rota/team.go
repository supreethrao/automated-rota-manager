package rota

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
	"github.com/supreethrao/automated-rota-manager/pkg/keys"
	"github.com/supreethrao/automated-rota-manager/pkg/localdb"

	"gopkg.in/yaml.v2"
)

const (
	maxUInt16 = 65535
)

// name will be used as the key prefix
type Team struct {
	name string
	db *localdb.LocalDB
	keys.Keys
}

type outofoffice struct {
	Name        string
	OutOfOffice string
}

type teamMembers struct {
	Members []string `yaml:"members"`
}

func (t Team) List() ([]string, error) {
	data, err := t.db.Read(t.TeamKey())
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return []string{}, nil
		}
		return nil, err
	}

	members := teamMembers{}

	if err = yaml.Unmarshal(data, &members); err != nil {
		log.Panicf("Unable to obtain team members: %v", err)
	}

	return members.Members, nil
}

func (t Team) Add(newMember string) error {
	currentMembers, err := t.List()
	if err != nil {
		return err
	}

	for _, member := range currentMembers {
		if newMember == member {
			log.Printf("%s is already a member", newMember)
			return nil
		}
	}

	updatedTeam := teamMembers{append(currentMembers, newMember)}
	if data, err := yaml.Marshal(updatedTeam); err == nil {
		multiData := map[string][]byte{
			t.TeamKey():                        data,
			t.AccruedDaysCounterKey(newMember): uintToBytes(t.lowestAccruedDaysAmongstTeamMembers()),
		}
		return t.db.MultiWrite(multiData)
	} else {
		return err
	}
}

func (t Team) Remove(existingMember string) error {
	currentMembers, err := t.List()
	if err != nil {
		return err
	}
	updatedMembers := make([]string, 0)

	for _, member := range currentMembers {
		if member != existingMember {
			updatedMembers = append(updatedMembers, member)
		}
	}
	updatedTeam := teamMembers{updatedMembers}
	if data, err := yaml.Marshal(updatedTeam); err == nil {
		return t.db.Write(t.TeamKey(), data)
	} else {
		return err
	}
}

func (t Team) HistoryOfIndividual(member string) IndividualHistory {
	history := IndividualHistory{member, 0, "N/A"}
	count, err := t.db.Read(t.AccruedDaysCounterKey(member))
	if err == nil {
		history.DaysAccrued = bytesToUint(count)
	}

	day, err := t.db.Read(t.LatestDayPickedKey(member))
	if err == nil {
		history.LatestPickedDay = string(day)
	}

	return history
}

func (t Team) RotaHistory() (TeamRotaHistory, error) {
	teamHistory := make([]IndividualHistory, 0)
	teamList, err := t.List()
	if err != nil {
		 return nil, err
	}
	for _, member := range teamList {
		teamHistory = append(teamHistory, t.HistoryOfIndividual(member))
	}
	return teamHistory, nil
}


func (t Team) SetOutOfOffice(memberName string, from time.Time, to time.Time) error {
	fromDate := from.Format("02-01-2006")
	toDate := to.Format("02-01-2006")

	fromKey, toKey := t.OutOfOfficeKey(memberName)

	return t.db.MultiWrite(map[string][]byte{
		fromKey: []byte(fromDate),
		toKey:   []byte(toDate),
	})
}

func (t Team) GetOutOfOffice(memberName string) []byte {
	return t.outOfOffice([]string{memberName})
}

func (t Team) GetTeamOutOfOffice() ([]byte, error) {
	teamList, err := t.List()
	if err != nil {
		return nil, err
	}

	return t.outOfOffice(teamList), nil
}

func (t Team) IsAvailable(memberName string) bool {
	today := time.Now().Format("02-01-2006")
	fromKey, toKey := t.OutOfOfficeKey(memberName)

	from, errFrom := t.db.Read(fromKey)
	to, errTo := t.db.Read(toKey)

	if errFrom == nil && errTo == nil {
		fromDate, _ := time.Parse("02-01-2006", string(from))
		toDate, _ := time.Parse("02-01-2006", string(to))
		presentDate, _ := time.Parse("02-01-2006", today)

		return presentDate.Before(fromDate) || presentDate.After(toDate)
	}

	if errFrom == badger.ErrKeyNotFound || errTo == badger.ErrKeyNotFound {
		log.Printf("No out of office dates registered for %s \n", memberName)
	}
	return true
}

func (t Team) Next() (string, error) {
	history, err := t.RotaHistory()
	if err != nil {
		return "", err
	}
	teamRotaHistory := orderedList(history)

	if teamRotaHistory.Len() < 1 {
		return "UNKNOWN-HISTORY", nil
	}

	// Days in between picking same person. Set at 2 times the frequency. i.e same person won't be picked before having picked at least 2 others
	minDaysInBetween := 0
	lastRun, err := t.db.Read(t.LatestCronRunKey())
	if err != nil {
		logrus.Errorf("unable to obtain last run time: %v", err)
	} else {
		minDaysInBetween, err = differenceBetweenDays(string(lastRun), today())
		if err != nil {
			logrus.Errorf("unable to obtain difference in number of days since the cron run. Defaulting to 0: %v", err)
		} else {
			// There should be at least 2 different picks before the same person is picked again.
			minDaysInBetween *= 2
		}
	}

	logrus.Infof("min days in between picks %v", minDaysInBetween)
	for _, individual := range teamRotaHistory {
		latestPickedDay := "31-12-2006"
		// If the person is newly added and has not been picked yet, this value will be N/A. Else that person is ripe to be picked next
		if individual.LatestPickedDay != "N/A" {
			latestPickedDay = individual.LatestPickedDay
		}

		if diffBetweenLastPick, err := differenceBetweenDays(latestPickedDay, today()); err == nil {
			if  diffBetweenLastPick > minDaysInBetween {
				probablePerson := individual.Name

				if t.IsAvailable(probablePerson) {
					return probablePerson, nil
				}
			}
		} else {
			return "UNKNOWN-ERROR", err
		}
	}
	return "UNKNOWN-UNKNOWN", nil
}

func (t Team) outOfOffice(names []string) []byte {
	var oooRecords []outofoffice

	for _, memberName := range names {
		fromKey, toKey := t.OutOfOfficeKey(memberName)

		from, errFrom := t.db.Read(fromKey)
		to, errTo := t.db.Read(toKey)

		if errFrom == nil && errTo == nil {
			oooRecords = append(oooRecords, outofoffice{memberName, "From " + string(from) + " To " + string(to)})
		}
	}

	if data, err := json.Marshal(&oooRecords); err != nil {
		return []byte("Unable to retrieve")
	} else {
		return data
	}
}

func uintToBytes(val uint16) []byte {
	bytesVal := make([]byte, 2)
	binary.BigEndian.PutUint16(bytesVal, val)
	return bytesVal
}

func bytesToUint(val []byte) uint16 {
	return binary.BigEndian.Uint16(val)
}

func differenceBetweenDays(ddmmyyyyStr1, ddmmyyyystr2 string) (int, error) {
	firstDay, e1 := time.Parse("02-01-2006", ddmmyyyyStr1)
	if e1 != nil {
		return 0, fmt.Errorf("Unable to parse date string %s - %v", ddmmyyyyStr1, e1)
	}
	secondDay, e2 := time.Parse("02-01-2006", ddmmyyyystr2)
	if e2 != nil {
		return 0, fmt.Errorf("Unable to parse date string %s - %v", ddmmyyyystr2, e2)
	}

	// check second day is after the first day
	if secondDay.After(firstDay) {
		return int(math.Round(math.Abs(secondDay.Sub(firstDay).Hours() / 24))), nil
	}
	return 0, fmt.Errorf("cron run date %v is in the future and this is incorrect. ", secondDay)
}

func today() string {
	return time.Now().Format("02-01-2006")
}

func (t Team) lowestAccruedDaysAmongstTeamMembers() uint16 {
	var lowestAccruedDays = uint16(maxUInt16)

	history, err := t.RotaHistory()
	if err != nil {
		return 1
	}

	for _, individualHistory := range history {
		if individualHistory.DaysAccrued < lowestAccruedDays {
			lowestAccruedDays = individualHistory.DaysAccrued
		}
	}

	// This conditional required while adding the very first team member on a new deployment
	if lowestAccruedDays == maxUInt16 {
		return 1
	}
	return lowestAccruedDays
}

func NewTeam(name string, dbHandle *localdb.LocalDB) *Team {
	return &Team{
		name,
		dbHandle,
		keys.NewKey(name),
	}
}
