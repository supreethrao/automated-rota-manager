package rota

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"github.com/sky-uk/support-bot/localdb"
	"github.com/sky-uk/support-bot/rota/keys"
	"gopkg.in/yaml.v2"
	"log"
	"time"
)

type Team interface {
	List() []string
	Add(newMember string) error
	AddTeam(teamMembers string) error
	Remove(existingMember string) error
	SupportHistoryOfIndividual(member string) IndividualSupportHistory
	SupportHistoryForTeam() TeamSupportHistory
	SupportPersonOnTheDay(date time.Time) string
	SetPersonOnSupportForToday(memberName string) error
	OverrideSupportPersonForToday(memberName string) error
	SetOutOfOffice(memberName string, from time.Time, to time.Time) error
	GetOutOfOffice(memberName string) []byte
	GetTeamOutOfOffice() []byte
	IsAvailable(name string) bool
	keys.Keys
}

// name will be used as the key prefix
type team struct {
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

func (t *team) List() []string {
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

func (t *team) Add(newMember string) error {
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
			t.SupportDaysCounterKey(newMember): uintToBytes(0),
		}
		return localdb.MultiWrite(multiData)
	} else {
		return err
	}
}

func (t *team) AddTeam(teamMembers string) error {
	return nil
}

func (t *team) Remove(existingMember string) error {
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

func (t *team) SupportHistoryOfIndividual(member string) IndividualSupportHistory {
	history := IndividualSupportHistory{member, 0, "31-12-2006"}
	count, err := localdb.Read(t.SupportDaysCounterKey(member))
	if err == nil {
		history.DaysSupported = bytesToUint(count)
	}

	day, err := localdb.Read(t.LatestDayOnSupportKey(member))
	if err == nil {
		history.LatestSupportDay = string(day)
	}

	return history
}

func (t *team) SupportHistoryForTeam() TeamSupportHistory {
	teamHistory := make([]IndividualSupportHistory, 0)
	for _, member := range t.List() {
		teamHistory = append(teamHistory, t.SupportHistoryOfIndividual(member))
	}
	return teamHistory
}

func (t *team) SupportPersonOnTheDay(date time.Time) string {
	supportPerson, err := localdb.Read(t.SupportPersonOnDayKey(date))
	if err == nil {
		return string(supportPerson)
	}

	log.Printf("Unable to retrieve support person for %v", date)
	return "UNKNOWN"
}

func (t *team) SetPersonOnSupportForToday(memberName string) error {
	supportKeys := make(map[string][]byte)

	personAssignedForTheDay := t.SupportPersonOnTheDay(time.Now())

	if personAssignedForTheDay != "UNKNOWN" {
		return errors.New(fmt.Sprintf("%s is already assigned for the day", personAssignedForTheDay))
	}

	currentlySupportedDays, _ := localdb.Read(t.SupportDaysCounterKey(memberName))
	incrementedSupportDays := uintToBytes(bytesToUint(currentlySupportedDays) + 1)

	supportKeys[t.SupportDaysCounterKey(memberName)] = incrementedSupportDays
	supportKeys[t.LatestDayOnSupportKey(memberName)] = []byte(today())
	supportKeys[t.SupportPersonOnDayKey(time.Now())] = []byte(memberName)

	return localdb.MultiWrite(supportKeys)
}

func (t *team) OverrideSupportPersonForToday(memberName string) error {
	supportKeys := make(map[string][]byte)

	personAssignedForTheDay := t.SupportPersonOnTheDay(time.Now())

	if personAssignedForTheDay == "UNKNOWN" {
		return t.SetPersonOnSupportForToday(memberName)
	}

	supportedDaysAsBytes, _ := localdb.Read(t.SupportDaysCounterKey(personAssignedForTheDay))
	adjustedSupportDays := uint16(0)
	supportedDays := bytesToUint(supportedDaysAsBytes)
	if supportedDays > 0 {
		adjustedSupportDays = supportedDays - 1
	}

	supportKeys[t.SupportDaysCounterKey(personAssignedForTheDay)] = uintToBytes(adjustedSupportDays)
	// This is incorrect - need to traverse through the history and get the date this person was previously on support
	supportKeys[t.LatestDayOnSupportKey(personAssignedForTheDay)] = []byte("31-12-2006")

	currentlySupportedDays, _ := localdb.Read(t.SupportDaysCounterKey(memberName))
	incrementedSupportDays := uintToBytes(bytesToUint(currentlySupportedDays) + 1)

	supportKeys[t.SupportDaysCounterKey(memberName)] = incrementedSupportDays
	supportKeys[t.LatestDayOnSupportKey(memberName)] = []byte(today())
	supportKeys[t.SupportPersonOnDayKey(time.Now())] = []byte(memberName)

	return localdb.MultiWrite(supportKeys)
}

func (t *team) SetOutOfOffice(memberName string, from time.Time, to time.Time) error {
	fromDate := from.Format("02-01-2006")
	toDate := to.Format("02-01-2006")

	fromKey, toKey := t.OutOfOfficeKey(memberName)

	return localdb.MultiWrite(map[string][]byte{
		fromKey: []byte(fromDate),
		toKey:   []byte(toDate),
	})
}

func (t *team) GetOutOfOffice(memberName string) []byte {
	return t.outOfOffice([]string{memberName})
}

func (t *team) GetTeamOutOfOffice() []byte {
	return t.outOfOffice(t.List())
}

func (t *team) IsAvailable(memberName string) bool {
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

func (t *team) outOfOffice(names []string) []byte {

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

func NewTeam(name string) Team {
	return &team{
		name,
		keys.NewKey(name),
	}
}
