package usageevents_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"os"
	"app-metrics-nozzle/api/apifakes"
	"app-metrics-nozzle/api"
	"github.com/cloudfoundry-community/go-cfclient"
	. "app-metrics-nozzle/usageevents"
	"app-metrics-nozzle/usageevents/usageeventsfakes"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
)

var _ = Describe("usageevents", func() {
	var (
		simpleApp cfclient.App
		space cfclient.Space
		org cfclient.Org
		allApps	[]caching.App
		appInstances map[string]cfclient.AppInstance
		spaceUsers []cfclient.User
		orgUsers []cfclient.User
		fakeClient *apifakes.FakeCFClientCaller
		fakeCaching *usageeventsfakes.FakeAppCaching
	)

	BeforeEach(func() {
		loadJsonFromFile("fixtures/returned_app.json", &simpleApp)
		loadJsonFromFile("fixtures/space_users.json", &spaceUsers)
		loadJsonFromFile("fixtures/org_users.json", &orgUsers)
		loadJsonFromFile("fixtures/app_space.json", &space)
		loadJsonFromFile("fixtures/space_org.json", &org)
		fakeClient = new(apifakes.FakeCFClientCaller)
		fakeClient.UsersByReturns(spaceUsers, nil)
		fakeClient.AppByGuidReturns(simpleApp, nil)
		fakeClient.UsersByReturns(orgUsers, nil)
		fakeClient.AppSpaceReturns(space, nil)
		fakeClient.SpaceOrgReturns(org, nil)

		loadJsonFromFile("fixtures/all_cached_apps.json", &allApps)
		fakeCaching = new(usageeventsfakes.FakeAppCaching)
		fakeCaching.GetAllAppReturns(allApps)
		fakeCaching.GetAppInfoReturns(allApps[0])
	})

	Describe("Given: a ReloadApps function", func() {
		Context("When: called on a Appdetails not containing any apps", func() {
			BeforeEach(func() {
				loadJsonFromFile("fixtures/app_instances.json", &appInstances)
				fakeClient.GetAppInstancesReturns(appInstances, nil)
				api.Client = fakeClient
			})
			It("then: it should populate the appdetails objects with app info from Cloud Controller", func() {
				Expect(len(AppDetails)).To(Equal(0))
				ReloadApps(fakeCaching.GetAllApp())
				Expect(len(AppDetails)).To(BeNumerically(">", 0))
				Expect(AppDetails["system/system/apps-manager-js"].InstanceCount.Configured).To(BeNumerically("==", 6))
				Expect(AppDetails["system/system/apps-manager-js"].InstanceCount.Running).To(BeNumerically("==", 6))
				Expect(AppDetails["system/system/apps-manager-js"].Diego).To(Equal(true))
				Expect(len(AppDetails["system/system/apps-manager-js"].Routes)).To(Equal(2))
			})
		})
	})
})

func loadJsonFromFile(filePath string, obj interface{})  {
	file, e := ioutil.ReadFile(filePath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	json.Unmarshal(file, obj)
}

