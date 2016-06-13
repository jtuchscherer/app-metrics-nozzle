package api

import (
	"testing"
	"app-usage-nozzle/domain"
	"os"
	"github.com/jtgammon/go-cfclient"
	"fmt"
)


var AppDetails = make(map[string]domain.App)

func TestReverse(t *testing.T) {

	c := cfclient.Config{
		ApiAddress:        "https://api.run.haas-41.pez.pivotal.io",
		Username:          "admin",
		Password:          "cb0a40f8d6360eaed442",
		SkipSslValidation: true,
	}

	logger.Println("Processing Cloud Controller call to " + os.Getenv("API_ENDPOINT"))
	client, _ = cfclient.NewClient(&c)


	orgs, _ := client.ListOrgs()
	var o []Entity // == nil
	for idx := range orgs {
		org := Entity{Name:orgs[idx].Name, Guid:orgs[idx].Guid}
		o = append(o, org)
	}
	logger.Println(fmt.Sprintf("Org name %s", o))


	spaces, _ := client.ListSpaces()
	for idx := range spaces {
		org := Entity{Name:spaces[idx].Name, Guid:spaces[idx].Guid}
		o = append(o, org)
	}
	logger.Println(fmt.Sprintf("Space name %s", o))

}

type Entity struct {
	Name	string	`json:"name"`
	Guid	string	`json:"guid"`
}
