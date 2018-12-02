package rota_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/support-bot/localdb"
	"github.com/sky-uk/support-bot/rota"
	"github.com/sky-uk/support-bot/rota_test/helper"
	"testing"
	"time"
)

func TestTeam(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test suite for team")
}

var _ = Describe("CRUD of team members", func() {

	var myTeam rota.Team = nil
	BeforeSuite(func() {
		myTeam = rota.NewTeam("test_team")
	})

	BeforeEach(func() {
		Expect(localdb.Remove(myTeam.TeamKey())).To(Succeed())
		for _, member := range helper.TestTeamMembers {
			Expect(localdb.Remove(myTeam.SupportDaysCounterKey(member))).To(Succeed())
			Expect(localdb.Remove(myTeam.LatestDayOnSupportKey(member))).To(Succeed())
			Expect(localdb.Remove(myTeam.SupportPersonOnDayKey(time.Now()))).To(Succeed())
		}
		Expect(localdb.Write(myTeam.TeamKey(), helper.TestTeamMembersListYaml))
	})

	Context("Read team members", func() {
		It("List gets data from the members file", func() {
			Expect(myTeam.List()).To(Equal([] string{"person1", "person2", "third person"}))
		})
	})

	Context("Adding new team members", func() {
		It("Add new team member adds the member to the list", func() {
			Expect(myTeam.Add("new member")).To(Succeed())
			Expect(myTeam.List()).To(Equal([] string{"person1", "person2", "third person", "new member"}))
		})

		It("Add new team member should not fail if the member already exists", func() {
			personToAdd := "third person"
			Expect(myTeam.Add(personToAdd)).To(Succeed())
			Expect(myTeam.List()).To(Equal([] string{"person1", "person2", "third person"}))
		})

		It("Add new team member initialise their support counter key to 0", func() {
			newTeamMember := "fourth person"
			Expect(myTeam.Add(newTeamMember)).To(Succeed())
			Expect(myTeam.SupportHistoryOfIndividual(newTeamMember).DaysSupported).To(Equal(uint16(0)))
		})

		It("Adding an existing team member again should not reset the supported days ", func() {
			existingTeamMember := "person1"
			Expect(localdb.Write(myTeam.SupportDaysCounterKey(existingTeamMember), helper.Uint16ToBytes(7))).To(Succeed())

			Expect(myTeam.Add(existingTeamMember)).To(Succeed())
			Expect(myTeam.SupportHistoryOfIndividual(existingTeamMember).DaysSupported).To(Equal(uint16(7)))
		})
	})

	Context("Removing team members", func() {
		It("Removing existing team member returns success", func() {
			Expect(myTeam.Remove("third person")).To(Succeed())
			Expect(myTeam.List()).To(Equal([] string{"person1", "person2"}))
		})
		It("Removing non-existing team member returns success", func() {
			Expect(myTeam.Remove("non-existent person")).To(Succeed())
			Expect(myTeam.List()).To(Equal([] string{"person1", "person2", "third person"}))
		})
	})

	Context("Setting the person on support", func() {
		It("The person being set on support will have the relevant keys updated", func() {
			// given
			Expect(localdb.Write(myTeam.SupportDaysCounterKey("person1"), helper.Uint16ToBytes(7))).To(Succeed())

			//when
			Expect(myTeam.SetPersonOnSupportForToday("person1")).To(Succeed())

			//then
			Expect(localdb.Read(myTeam.SupportDaysCounterKey("person1"))).To(Equal(helper.Uint16ToBytes(8)))
			Expect(localdb.Read(myTeam.LatestDayOnSupportKey("person1"))).To(Equal([]byte(helper.Today())))
			Expect(localdb.Read(myTeam.SupportPersonOnDayKey(time.Now()))).To(Equal([]byte("person1")))
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

