package usageevents_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jtuchscherer/app-metrics-nozzle/usageevents/usageeventsfakes"

	"github.com/jtuchscherer/app-metrics-nozzle/domain"
	. "github.com/jtuchscherer/app-metrics-nozzle/usageevents"

	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/jtuchscherer/app-metrics-nozzle/api"
	"github.com/jtuchscherer/app-metrics-nozzle/api/apifakes"
)

var _ = Describe("usageevents", func() {
	var (
		simpleApp    cfclient.App
		rtrEvent     events.Envelope
		metricsEvent events.Envelope
		space        cfclient.Space
		org          cfclient.Org
		allApps      []caching.App
		appInstances map[string]cfclient.AppInstance
		fakeClient   *apifakes.FakeCFClientCaller
		fakeCaching  *usageeventsfakes.FakeCachedApp

		testAppKey   string
		testAppKeyCC string
	)

	BeforeEach(func() {
		testAppKey = "Pivotal/ashumilov/cd-demo-music"
		testAppKeyCC = "system/system/apps-manager-js"
		loadJsonFromFile("fixtures/rtr_log_message.json", &rtrEvent)
		loadJsonFromFile("fixtures/container_metric_log_message.json", &metricsEvent)

		loadJsonFromFile("fixtures/returned_app.json", &simpleApp)
		loadJsonFromFile("fixtures/app_space.json", &space)
		loadJsonFromFile("fixtures/space_org.json", &org)
		fakeClient = new(apifakes.FakeCFClientCaller)
		fakeClient.AppByGuidReturns(simpleApp, nil)
		fakeClient.AppSpaceReturns(space, nil)
		fakeClient.SpaceOrgReturns(org, nil)

		loadJsonFromFile("fixtures/all_cached_apps.json", &allApps)
		fakeCaching = new(usageeventsfakes.FakeCachedApp)
		fakeCaching.GetAllAppReturns(allApps)
		fakeCaching.GetAppInfoReturns(allApps[11])
	})

	Describe("Given: a Firehouse events", func() {
		BeforeEach(func() {
			AppDetails = make(map[string]domain.App)
			loadJsonFromFile("fixtures/app_instances.json", &appInstances)
			fakeClient.GetAppInstancesReturns(appInstances, nil)
			api.Client = fakeClient
			AppDbCache = fakeCaching

			Expect(len(AppDetails)).To(Equal(0))
			ReloadApps(fakeCaching.GetAllApp())
		})
		Context("When: processed Cloud Controller call", func() {
			It("then: it should populate the appdetails objects with app info from data returned from CC", func() {
				Expect(len(AppDetails)).To(BeNumerically(">", 0))
				Expect(AppDetails[testAppKeyCC].InstanceCount.Configured).To(BeNumerically("==", 6))
				Expect(AppDetails[testAppKeyCC].InstanceCount.Running).To(BeNumerically("==", 6))
				Expect(AppDetails[testAppKeyCC].Diego).To(Equal(true))
				Expect(len(AppDetails[testAppKeyCC].Routes)).To(Equal(3))
			})
		})
		Context("When: processed RTR event", func() {
			It("then: it should populate the appdetails objects with app info from event with source type RTR", func() {
				ProcessEvent(&rtrEvent)
				Expect(len(AppDetails)).To(BeNumerically(">", 0))
				Expect(AppDetails[testAppKey].EventCount).To(BeNumerically("==", 1))
				Expect(AppDetails[testAppKey].LastEventTime).ToNot(BeNil())
			})
		})
		Context("When: processed app metrics event", func() {
			It("then: it should populate the appdetails objects with app info from application metrics event", func() {
				ProcessEvent(&metricsEvent)
				Expect(AppDetails[testAppKey].Instances[5].CellIP).ToNot(BeNil())
				Expect(AppDetails[testAppKey].Instances[5].CPUUsage).ToNot(BeNil())
				Expect(AppDetails[testAppKey].Instances[5].DiskUsage).ToNot(BeNil())
				Expect(AppDetails[testAppKey].Instances[5].MemoryUsage).ToNot(BeNil())
			})
		})

	})
})

func loadJsonFromFile(filePath string, obj interface{}) {
	file, e := ioutil.ReadFile(filePath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	json.Unmarshal(file, obj)
}
