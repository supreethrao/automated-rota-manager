package rota_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/supreethrao/support-bot/rota"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"testing"
)

func TestRota(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test suite for team")
}

var _ = Describe("Tests retrieving team", func() {
	Context("Initialise data", func() {
		FIt("New gets data from the file", func() {
			Expect(rota.List()).To(Equal([] string{"someone", "someone else", "some random person"}))
		})
	})

	Context("Test writing data back to a file", func() {
		It("Create a struct and write as a yaml file", func() {
			type team struct {
				Members []string `yaml:"members"`
			}

			myteam := team{[]string{"hello", "world"}}
			data, err := yaml.Marshal(myteam)

			if err == nil {
				damn := ioutil.WriteFile("/tmp/test-team.yml", data, os.ModePerm)
				if damn != nil {
					fmt.Println("DAMN")
				}
			} else {
				fmt.Println("SUCKS")
			}
		})
	})
})
