package rota_test

import (
	"github.com/supreethrao/automated-rota-manager/pkg/rota"
	"time"

	"github.com/supreethrao/automated-rota-manager/pkg/localdb"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CRUD of team members", func() {

	var myTeam *rota.Team
	BeforeSuite(func() {
		myTeam = rota.NewTeam("test_team")
	})

	BeforeEach(func() {
		Expect(localdb.Remove(myTeam.TeamKey())).To(Succeed())
		for _, member := range testTeamMembers {
			Expect(localdb.Remove(myTeam.AccruedDaysCounterKey(member))).To(Succeed())
			Expect(localdb.Remove(myTeam.LatestDayPickedKey(member))).To(Succeed())
			Expect(localdb.Remove(myTeam.PersonPickedOnDayKey(time.Now()))).To(Succeed())
		}
		Expect(localdb.Write(myTeam.TeamKey(), TestTeamMembersListYaml))
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
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey(existingTeamMember), Uint16ToBytes(7))).To(Succeed())

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
			Expect(localdb.Write(myTeam.AccruedDaysCounterKey("person1"), Uint16ToBytes(7))).To(Succeed())

			//when
			Expect(myTeam.SetPersonPickedForToday("person1")).To(Succeed())

			//then
			Expect(localdb.Read(myTeam.AccruedDaysCounterKey("person1"))).To(Equal(Uint16ToBytes(8)))
			Expect(localdb.Read(myTeam.LatestDayPickedKey("person1"))).To(Equal([]byte(Today())))
			Expect(localdb.Read(myTeam.PersonPickedOnDayKey(time.Now()))).To(Equal([]byte("person1")))
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
})

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
