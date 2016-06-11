package api

import (
	cfclient "github.com/jtgammon/go-cfclient"
	"os"
	"log"
	"strconv"
	"strings"
	"app-usage-nozzle/domain"
)

var logger = log.New(os.Stdout, "", 0)
var client *cfclient.Client

func init(){
	c := cfclient.Config{
		ApiAddress:        "https://api.run.haas-41.pez.pivotal.io",
		Username:          "admin",
		Password:          "cb0a40f8d6360eaed442",
		SkipSslValidation: true,
	}
	client, _ = cfclient.NewClient(&c)
}

func AnnotateWithCloudControllerData(app *domain.App) {

	ccAppDetails, _ := client.AppByGuid(app.GUID)

	instances, _ := client.GetAppInstances(app.GUID)
	runnintCount := 0
	instanceUp := "RUNNING"

	space, _ := ccAppDetails.Space()
	org, _ := space.Org()

	app.Diego = ccAppDetails.Diego
	app.Buildpack = ccAppDetails.Buildpack
	app.Instances = make([]domain.Instances, int64(len(instances)))

	for idx, eachInstance := range instances {
		if strings.Compare(instanceUp, eachInstance.State) == 0 {
			runnintCount++;
		}
		i, _ := strconv.ParseInt(idx, 10, 32)
		app.Instances[i].InstanceIndex = i
		app.Instances[i].State = eachInstance.State
		app.Instances[i].Since = eachInstance.Since
		app.Instances[i].Uptime = eachInstance.Uptime
	}

	if len(app.Buildpack) == 0 { app.Buildpack = ccAppDetails.DetectedBP }

	app.Environment = ccAppDetails.Environment

	app.Organization.ID = org.Guid
	app.Organization.Name = org.Name

	app.Space.ID = space.Guid
	app.Space.Name = space.Name

	app.InstanceCount.Configured = len(instances)
	app.InstanceCount.Running = runnintCount

	app.EnvironmentSummary.TotalDiskConfigured = ccAppDetails.DiskQuota
	app.EnvironmentSummary.TotalMemoryConfigured = ccAppDetails.MemQuota

	app.EnvironmentSummary.TotalDiskProvisioned = ccAppDetails.DiskQuota * int32(len(instances))
	app.EnvironmentSummary.TotalMemoryProvisioned = ccAppDetails.MemQuota * int32(len(instances))

	if 0 < len(ccAppDetails.RouteData) {
		app.Routes = make([]string, len(ccAppDetails.RouteData))
		for i := 0; i < len(ccAppDetails.RouteData); i++ {
			app.Routes[i] = ccAppDetails.RouteData[i].Entity.Host + "." + ccAppDetails.RouteData[i].Entity.DomainData.Entity.Name
		}
	}

	app.State = ccAppDetails.State
}



