package rota

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"time"

	"github.com/supreethrao/automated-rota-manager/pkg/keys"
	"github.com/supreethrao/automated-rota-manager/pkg/localdb"

	"github.com/dgraph-io/badger"
	"gopkg.in/yaml.v2"
)

// name will be used as the key prefix
type Team struct {
	name string
	keys.Keys
}

type outofoffice struct {
	Name        string
	OutOfOffice string
}

type teamMembers struct {
	Members []string `yaml:"members"`
}

func (t Team) List() []string {
	data, err := localdb.Read(t.TeamKey())
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return []string{}
		}
		panic(err)
	}

	members := teamMembers{}

	if err = yaml.Unmarshal(data, &members); err != nil {
		log.Panicf("Unable to obtain team members: %v", err)
	}

	return members.Members
}

func (t Team) Add(newMember string) error {
	currentMembers := t.List()

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
			t.AccruedDaysCounterKey(newMember): uintToBytes(0),
		}
		return localdb.MultiWrite(multiData)
	} else {
		return err
	}
}

func (t Team) Remove(existingMember string) error {
	currentMembers := t.List()
	updatedMembers := make([]string, 0)

	for _, member := range currentMembers {
		if member != existingMember {
			updatedMembers = append(updatedMembers, member)
		}
	}
	updatedTeam := teamMembers{updatedMembers}
	if data, err := yaml.Marshal(updatedTeam); err == nil {
		return localdb.Write(t.TeamKey(), data)
	} else {
		return err
	}
}

func (t Team) HistoryOfIndividual(member string) IndividualHistory {
	history := IndividualHistory{member, 0, "31-12-2006"}
	count, err := localdb.Read(t.AccruedDaysCounterKey(member))
	if err == nil {
		history.DaysAccrued = bytesToUint(count)
	}

	day, err := localdb.Read(t.LatestDayPickedKey(member))
	if err == nil {
		history.LatestPickedDay = string(day)
	}

	return history
}

func (t Team) RotaHistory() TeamRotaHistory {
	teamHistory := make([]IndividualHistory, 0)
	for _, member := range t.List() {
		teamHistory = append(teamHistory, t.HistoryOfIndividual(member))
	}
	return teamHistory
}


func (t Team) SetOutOfOffice(memberName string, from time.Time, to time.Time) error {
	fromDate := from.Format("02-01-2006")
	toDate := to.Format("02-01-2006")

	fromKey, toKey := t.OutOfOfficeKey(memberName)

	return localdb.MultiWrite(map[string][]byte{
		fromKey: []byte(fromDate),
		toKey:   []byte(toDate),
	})
}

func (t Team) GetOutOfOffice(memberName string) []byte {
	return t.outOfOffice([]string{memberName})
}

func (t Team) GetTeamOutOfOffice() []byte {
	return t.outOfOffice(t.List())
}

func (t Team) IsAvailable(memberName string) bool {
	today := time.Now().Format("02-01-2006")
	fromKey, toKey := t.OutOfOfficeKey(memberName)

	from, errFrom := localdb.Read(fromKey)
	to, errTo := localdb.Read(toKey)

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

func (t Team) Next() string {
	teamRotaHistory := orderedList(t.RotaHistory())

	if teamRotaHistory.Len() < 1 {
		return "UNKNOWN"
	}

	for _, individual := range teamRotaHistory {
		if differenceBetweenDays(individual.LatestPickedDay, today()) > 2 {
			probablePerson := individual.Name

			if t.IsAvailable(probablePerson) {
				return probablePerson
			}
		}
	}
	return "UNKNOWN"
}

func (t Team) outOfOffice(names []string) []byte {

	var oooRecords []outofoffice

	for _, memberName := range names {
		fromKey, toKey := t.OutOfOfficeKey(memberName)

		from, errFrom := localdb.Read(fromKey)
		to, errTo := localdb.Read(toKey)

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

func differenceBetweenDays(ddmmyyyyStr1, ddmmyyyystr2 string) float64 {
	firstDay, e1 := time.Parse("02-01-2006", ddmmyyyyStr1)
	if e1 != nil {
		log.Panicf("Unable to parse date string %s - %v", ddmmyyyyStr1, e1)
	}
	secondDay, e2 := time.Parse("02-01-2006", ddmmyyyystr2)
	if e2 != nil {
		log.Panicf("Unable to parse date string %s - %v", ddmmyyyystr2, e2)
	}
	return math.Round(math.Abs(secondDay.Sub(firstDay).Hours() / 24))
}

func today() string {
	return time.Now().Format("02-01-2006")
}

func NewTeam(name string) *Team {
	return &Team{
		name,
		keys.NewKey(name),
	}
}
