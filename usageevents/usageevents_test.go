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
)

var _ = Describe("usageevents", func() {
	var (
		simpleApp    cfclient.App
		rtrEvent     events.Envelope
		metricsEvent events.Envelope
		space        cfclient.Space
		org          cfclient.Org
		allApps      []caching.App
		fakeClient   *usageeventsfakes.FakeApiClient
		fakeCaching  *usageeventsfakes.FakeCachedApp

		testAppKey   string
		testAppKeyCC string
	)

	BeforeEach(func() {
		fakeClient = &usageeventsfakes.FakeApiClient{}
		fakeClient.AnnotateWithCloudControllerDataStub = func(app *domain.App) {
			app.InstanceCount = domain.InstanceCount{
				Running:    6,
				Configured: 6,
			}
			app.Diego = true
			app.Routes = []string{"bla.test.com", "bla.test.org", "bla.test.net"}
			app.Instances = []domain.Instance{domain.Instance{}, domain.Instance{}, domain.Instance{
				CellIP:      "sdfsdf-sdfsdf",
				CPUUsage:    0.45,
				DiskUsage:   5634,
				MemoryUsage: 10345,
			}}
		}
		testAppKey = "Pivotal/ashumilov/cd-demo-music"
		testAppKeyCC = "system/system/apps-manager-js"
		loadJsonFromFile("fixtures/rtr_log_message.json", &rtrEvent)
		loadJsonFromFile("fixtures/container_metric_log_message.json", &metricsEvent)

		loadJsonFromFile("fixtures/returned_app.json", &simpleApp)
		loadJsonFromFile("fixtures/app_space.json", &space)
		loadJsonFromFile("fixtures/space_org.json", &org)

		loadJsonFromFile("fixtures/all_cached_apps.json", &allApps)
		fakeCaching = new(usageeventsfakes.FakeCachedApp)
		fakeCaching.GetAllAppReturns(allApps)
		fakeCaching.GetAppInfoReturns(allApps[11])
	})

	Describe("Given: a Firehouse events", func() {
		BeforeEach(func() {
			AppDetails = make(map[string]domain.App)

			AppDbCache = fakeCaching

			Expect(len(AppDetails)).To(Equal(0))
			ReloadApps(fakeCaching.GetAllApp(), fakeClient)
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
				Expect(AppDetails[testAppKey].Instances[2].CellIP).ToNot(BeNil())
				Expect(AppDetails[testAppKey].Instances[2].CPUUsage).ToNot(BeNil())
				Expect(AppDetails[testAppKey].Instances[2].DiskUsage).ToNot(BeNil())
				Expect(AppDetails[testAppKey].Instances[2].MemoryUsage).ToNot(BeNil())
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
