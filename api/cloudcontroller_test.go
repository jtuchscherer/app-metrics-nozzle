package api

import (
	"testing"
	//"github.com/davecgh/go-spew/spew"

	"github.com/cloudfoundry-community/go-cfclient"
	//"os"
	//"bufio"
	"encoding/json"
	"os"
	"bufio"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/boltdb/bolt"
	"time"
)

var client *cfclient.Client

func TestSuite(t *testing.T) {

	c := cfclient.Config{
		ApiAddress:        "https://api.run.haas-41.pez.pivotal.io",
		Username:          "admin",
		Password:          "cb0a40f8d6360eaed442",
		SkipSslValidation: true,
	}


	client, _ = cfclient.NewClient(&c)

	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		logger.Fatal("Error opening bolt db: ", err)
		os.Exit(1)

	}

	defer db.Close()

	caching.SetCfClient(client)
	caching.SetAppDb(db)
	caching.CreateBucket()


	allCachedApps := caching.GetAllApp()
	f, _ := os.Create("/Users/ashumilov/go/src/app-metrics-nozzle/usageevents/fixtures/all_cached_apps.json")
	w := bufio.NewWriter(f)
	a, _ := json.Marshal(allCachedApps)

	w.WriteString(string(a))

	f.Sync()
	w.Flush()


	//app, _ := client.AppByGuid("bb7b3c89-0a7f-47f7-9dd3-5e4fbd8ded6c")
	//
	//space, _ := app.Space()
	//org, _:= client.SpaceOrg(space)
	//
	//f, _ := os.Create("/Users/ashumilov/go/src/app-metrics-nozzle/api/fixtures/space_org.json")
	//w := bufio.NewWriter(f)
	//a, _ := json.Marshal(org)
	//
	//w.WriteString(string(a))
	//
	//f.Sync()
	//w.Flush()


	//appInstances, _ := client.GetAppInstances("0bcdb8a0-caa2-4db4-98c8-65f58d20a1d0")
	//
	//f, _ := os.Create("/Users/ashumilov/go/src/app-metrics-nozzle/usageevents/fixtures/app_instances.json")
	//w := bufio.NewWriter(f)
	//a, _ := json.Marshal(appInstances)
	//
	//w.WriteString(string(a))
	//
	//f.Sync()
	//w.Flush()


	//users, _  := client.UsersBy("dc4d1d1f-f4b9-4c60-8cbb-5763491d00c1", "spaces")
	//
	//f, _ := os.Create("/Users/ashumilov/go/src/app-metrics-nozzle/usageevents/fixtures/space_users.json")
	//w := bufio.NewWriter(f)
	//a, _ := json.Marshal(users)
	//
	//w.WriteString(string(a))

	//users, _  := client.UsersBy("c661e8c6-649a-4fe0-b471-afe5982e4e53", "organizations")
	//
	//f, _ := os.Create("/Users/ashumilov/go/src/app-metrics-nozzle/usageevents/fixtures/org_users.json")
	//w := bufio.NewWriter(f)
	//a, _ := json.Marshal(users)
	//
	//w.WriteString(string(a))


	//apps, _  := client.ListApps()
	//
	//f, _ := os.Create("/Users/ashumilov/go/src/app-metrics-nozzle/usageevents/fixtures/all_apps.json")
	//w := bufio.NewWriter(f)
	//a, _ := json.Marshal(apps)
	//
	//w.WriteString(string(a))


	//f, _ := os.Create("/Users/ashumilov/go/src/app-metrics-nozzle/usageevents/fixtures/app_space.json")
	//w := bufio.NewWriter(f)
	//s, _ := app.Space()
	//a, _ := json.Marshal(s)
	//
	//w.WriteString(string(a))
	//
	//f.Sync()
	//w.Flush()



}
