/*
Copyright 2016 Pivotal

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package api

import (
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"os"
	"log"
	"strconv"
	"strings"
	"app-metrics-nozzle/domain"
)

var logger = log.New(os.Stdout, "", 0)
var client *cfclient.Client

func init(){

	skipSsl, _ := strconv.ParseBool(os.Getenv("SKIP_SSL_VALIDATION"))
	c := cfclient.Config{
		ApiAddress:        os.Getenv("API_ENDPOINT"),
		Username:          os.Getenv("FIREHOSE_USER"),
		Password:          os.Getenv("FIREHOSE_PASSWORD"),
		SkipSslValidation: skipSsl,
	}

	logger.Println("Processing Cloud Controller call to " + os.Getenv("API_ENDPOINT"))
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

	app.EnvironmentSummary.TotalDiskConfigured = ccAppDetails.DiskQuota * 1024 * 1024
	app.EnvironmentSummary.TotalMemoryConfigured = ccAppDetails.MemQuota * 1024 * 1024

	app.EnvironmentSummary.TotalDiskProvisioned = ccAppDetails.DiskQuota * 1024 * 1024 * int32(len(instances))
	app.EnvironmentSummary.TotalMemoryProvisioned = ccAppDetails.MemQuota * 1024 * 1024 * int32(len(instances))

	if 0 < len(ccAppDetails.RouteData) {
		app.Routes = make([]string, len(ccAppDetails.RouteData))
		for i := 0; i < len(ccAppDetails.RouteData); i++ {
			app.Routes[i] = ccAppDetails.RouteData[i].Entity.Host + "." + ccAppDetails.RouteData[i].Entity.DomainData.Entity.Name
		}
	}

	app.State = ccAppDetails.State
}

func UsersForSpace(guid string) (Users []cfclient.User) {
	users, _ := client.UsersBy(guid, "spaces")
	return users
}

func UsersForOrganization(guid string) (Users []cfclient.User) {
	users, _ := client.UsersBy(guid, "organizations")
	return users
}

func SpacesDetailsFromCloudController()  (Spaces []cfclient.Space){
	spaces, _ := client.ListSpaces()
	return spaces
}

func OrgsDetailsFromCloudController()  (Orgs []cfclient.Org){
	orgs, _ := client.ListOrgs()
	return orgs
}




