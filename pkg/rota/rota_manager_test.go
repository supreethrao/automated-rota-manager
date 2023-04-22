package rota_test

import (
	"encoding/binary"
	"sort"
	"testing"
	"time"

	"github.com/supreethrao/automated-rota-manager/pkg/localdb"
	"github.com/supreethrao/automated-rota-manager/pkg/rota"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRotaLogic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Suite for rota manager")
}

var _ = Describe("Test suite for logic of picking next", func() {

	var dbHandle *localdb.LocalDB
	var myTeam *rota.Team

	BeforeSuite(func() {
		dbh, err := localdb.GetHandleFromLocation("/tmp/data")
		Expect(err).ToNot(HaveOccurred())

		dbHandle = dbh
	})

	BeforeEach(func() {
		myTeam = rota.NewTeam("test_team", dbHandle)
		Expect(dbHandle.Remove(myTeam.TeamKey())).To(Succeed())
		for _, member := range testTeamMembers {
			Expect(dbHandle.Remove(myTeam.AccruedDaysCounterKey(member))).To(Succeed())
			Expect(dbHandle.Remove(myTeam.PersonPickedOnDayKey(time.Now()))).To(Succeed())
			Expect(dbHandle.Remove(myTeam.LatestDayPickedKey(member))).To(Succeed())
			oooFrom, oooTo := myTeam.OutOfOfficeKey(member)
			Expect(dbHandle.Remove(oooFrom)).To(Succeed())
			Expect(dbHandle.Remove(oooTo)).To(Succeed())
		}
		Expect(dbHandle.Write(myTeam.TeamKey(), TestTeamMembersListYaml))
	})

	Context("Read team members", func() {
		It("List gets data from the members file", func() {
			Expect(myTeam.List()).To(Equal([]string{"person1", "person2", "third person"}))
		})
	})

	Context("Adding new team members", func() {
		It("Add new team member adds the member to the list", func() {
			Expect(myTeam.Add("new member")).To(Succeed())
			Expect(myTeam.List()).To(Equal([]string{"person1", "person2", "third person", "new member"}))
		})

		It("Add new team member should not fail if the member already exists", func() {
			personToAdd := "third person"
			Expect(myTeam.Add(personToAdd)).To(Succeed())
			Expect(myTeam.List()).To(Equal([]string{"person1", "person2", "third person"}))
		})

		It("Add new team member initialise their accrued days counter key to 0", func() {
			newTeamMember := "fourth person"
			Expect(myTeam.Add(newTeamMember)).To(Succeed())
			Expect(myTeam.HistoryOfIndividual(newTeamMember).DaysAccrued).To(Equal(uint16(0)))
		})

		It("Adding an existing team member again should not reset the accrued days ", func() {
			existingTeamMember := "person1"
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey(existingTeamMember), Uint16ToBytes(7))).To(Succeed())

			Expect(myTeam.Add(existingTeamMember)).To(Succeed())
			Expect(myTeam.HistoryOfIndividual(existingTeamMember).DaysAccrued).To(Equal(uint16(7)))
		})
	})

	Context("Removing team members", func() {
		It("Removing existing team member returns success", func() {
			Expect(myTeam.Remove("third person")).To(Succeed())
			Expect(myTeam.List()).To(Equal([]string{"person1", "person2"}))
		})
		It("Removing non-existing team member returns success", func() {
			Expect(myTeam.Remove("non-existent person")).To(Succeed())
			Expect(myTeam.List()).To(Equal([]string{"person1", "person2", "third person"}))
		})
	})

	Context("Setting the person picked", func() {
		It("The person picked will have the relevant keys updated", func() {
			// given
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(7))).To(Succeed())

			//when
			Expect(myTeam.SetPersonPickedForToday("person1")).To(Succeed())

			//then
			Expect(dbHandle.Read(myTeam.AccruedDaysCounterKey("person1"))).To(Equal(Uint16ToBytes(8)))
			Expect(dbHandle.Read(myTeam.LatestDayPickedKey("person1"))).To(Equal([]byte(Today())))
			Expect(dbHandle.Read(myTeam.PersonPickedOnDayKey(time.Now()))).To(Equal([]byte("person1")))
		})
	})

	Context("Batch add of team members or initialise the whole team", func() {
		It("Creates the entire team members from scratch", func() {

		})

		It("Adds multiple team members retaining old team members", func() {

		})

		It("Adding multiple team members don't add duplicates", func() {

		})
	})

	Context("Test sorting logic", func() {
		It("Should be sorted based on the accrued days", func() {
			teamHistory := rota.TeamRotaHistory{
				{"person1", 5, Yesterday()},
				{"person2", 3, Yesterday()},
				{"person3", 7, Yesterday()},
				{"person4", 2, Yesterday()},
			}

			expectedTeamHistory := rota.TeamRotaHistory{
				{"person4", 2, Yesterday()},
				{"person2", 3, Yesterday()},
				{"person1", 5, Yesterday()},
				{"person3", 7, Yesterday()},
			}

			sort.Sort(teamHistory)
			Expect(teamHistory).To(Equal(expectedTeamHistory))
		})
	})

	Context("Test picking based on fair rotation", func() {
		It("Next person is the person who has been fewer accrued days", func() {
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			nextPerson, err := myTeam.Next()
			Expect(err).ToNot(HaveOccurred())
			Expect(nextPerson).To(Equal("third person"))
		})

		It("Should have a couple of days breather regardless of number of accrued days", func() {
			Expect(dbHandle.Write(myTeam.LatestCronRunKey(), []byte(Yesterday())))

			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("person1"), []byte(Yesterday()))).To(Succeed())
			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("person2"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("third person"), []byte(DayBeforeYesterday()))).To(Succeed())

			nextPerson, err := myTeam.Next()
			Expect(err).ToNot(HaveOccurred())
			Expect(nextPerson).To(Equal("person2"))
		})
	})

	Context("Skip people who are out of office", func() {
		It("Skip the selected person if they are out of office", func() {
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			oooFrom, oooTo := myTeam.OutOfOfficeKey("third person")

			Expect(dbHandle.Write(oooFrom, timeToBytes(time.Now().Add(-time.Hour*24))))
			Expect(dbHandle.Write(oooTo, timeToBytes(time.Now().Add(time.Hour*24))))

			Expect(myTeam.Next()).To(Equal("person1"))
		})

		It("Skip the selected person who is off for the day", func() {
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(4))).To(Succeed())
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("person2"), Uint16ToBytes(6))).To(Succeed())
			Expect(dbHandle.Write(myTeam.AccruedDaysCounterKey("third person"), Uint16ToBytes(3))).To(Succeed())

			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("person1"), []byte(DaysBeforeToday(3)))).To(Succeed())
			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("person2"), []byte(DaysBeforeToday(4)))).To(Succeed())
			Expect(dbHandle.Write(myTeam.LatestDayPickedKey("third person"), []byte(DaysBeforeToday(5)))).To(Succeed())

			oooFrom, oooTo := myTeam.OutOfOfficeKey("third person")

			Expect(dbHandle.Write(oooFrom, timeToBytes(time.Now())))
			Expect(dbHandle.Write(oooTo, timeToBytes(time.Now())))

			Expect(myTeam.Next()).To(Equal("person1"))
		})
	})
})

func Today() string {
	return time.Now().Format("02-01-2006")
}

func Yesterday() string {
	return time.Now().AddDate(0, 0, -1).Format("02-01-2006")
}

func DaysBeforeToday(num int) string {
	return time.Now().AddDate(0, 0, -num).Format("02-01-2006")
}

func DayBeforeYesterday() string {
	return time.Now().AddDate(0, 0, -2).Format("02-01-2006")
}

func timeToBytes(t time.Time) []byte {
	return []byte(t.Format("02-01-2006"))
}

func Uint16ToBytes(intVal uint16) []byte {
	byteVal := make([]byte, 2)
	binary.BigEndian.PutUint16(byteVal, intVal)
	return byteVal
}

type testTeamMembersYaml struct {
	Members []string `yaml:"members"`
}

var testTeamMembers = []string{"person1", "person2", "third person"}
var TestTeamMembersListYaml = func() []byte {
	if yml, err := yaml.Marshal(testTeamMembersYaml{testTeamMembers}); err == nil {
		return yml
	} else {
		panic(err)
	}
}()