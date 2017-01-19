package usageevents

import (
	"fmt"

	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/jtuchscherer/app-metrics-nozzle/domain"
)

type apiClient interface {
	AnnotateWithCloudControllerData(app *domain.App)
}

// ReloadApps responsilbe for refreshing apps in the cache
func ReloadApps(cachedApps []caching.App, client apiClient) {
	logger.Println("Start filling app/space/org cache.")
	for idx := range cachedApps {

		org := cachedApps[idx].OrgName
		space := cachedApps[idx].SpaceName
		app := cachedApps[idx].Name
		key := GetMapKeyFromAppData(org, space, app)

		appId := cachedApps[idx].Guid
		name := cachedApps[idx].Name

		appDetail := &domain.App{GUID: appId, Name: name}
		client.AnnotateWithCloudControllerData(appDetail)
		AppDetails[key] = *appDetail
		logger.Println(fmt.Sprintf("Registered [%s]", key))
	}

	logger.Println(fmt.Sprintf("Done filling cache! Found [%d] Apps", len(cachedApps)))
}
