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

package usageevents

import (
	"fmt"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"sync"
	"time"
	"github.com/cloudfoundry/sonde-go/events"
	"strings"
	"strconv"
	"app-usage-nozzle/domain"
	"os"
	"log"
	"github.com/cloudfoundry-community/go-cfclient"
)

// Event is a struct represented an event augmented/decorated with corresponding app/space/org data.
type Event struct {
	Msg            string `json:"message"`
	Type           string `json:"event_type"`
	Origin         string `json:"origin"`
	AppID          string `json:"app_id"`
	Timestamp      int64  `json:"timestamp"`
	SourceType     string `json:"source_type"`
	MessageType    string `json:"message_type"`
	SourceInstance string `json:"source_instance"`
	AppName        string `json:"app_name"`
	OrgName        string `json:"org_name"`
	SpaceName      string `json:"space_name"`
	OrgID          string `json:"org_id"`
	SpaceID        string `json:"space_id"`
	CellIP         string `json:"cell_ip"`
	InstanceIndex  int32  `json:"instance_index"`
	CPUPercentage  float64 `json:"cpu_percentage"`
	MemBytes       uint64 `json:"mem_bytes"`
	DiskBytes      uint64 `json:"disk_bytes"`
}

var mutex sync.Mutex

var logger = log.New(os.Stdout, "", 0)

var AppDetails = make(map[string]domain.App)
var OrganizationUsers = make(map[string][]cfclient.User)
var SpacesUsers = make(map[string][]cfclient.User)
var Orgs []cfclient.Org
var Spaces []cfclient.Space

var feedStarted int64

// ProcessEvents churns through the firehose channel, processing incoming events.
func ProcessEvents(in <-chan *events.Envelope) {
	feedStarted = time.Now().UnixNano()
	for msg := range in {
		processEvent(msg)
	}
}

func processEvent(msg *events.Envelope) {
	eventType := msg.GetEventType()

	var event Event
	if eventType == events.Envelope_LogMessage {
		event = LogMessage(msg)
		if event.SourceType == "RTR" {
			event.AnnotateWithAppData()
			updateAppDetails(event)
		}

		if event.SourceType == "APP" {
			event.AnnotateWithAppData()
			updateAppWithAppEvent(event)
		}
	}

	if eventType == events.Envelope_ContainerMetric {
		event = ContainerMetric(msg)
		event.AnnotateWithAppData()
		updateAppWithContainerMetrics(event)
	}
}

func ContainerMetric(msg *events.Envelope) Event {
	message := msg.GetContainerMetric()

	return Event{
		Origin:         msg.GetOrigin(),
		Type:           msg.GetEventType().String(),
		AppID:          message.GetApplicationId(),
		CellIP:                *msg.Ip,
		InstanceIndex:  message.GetInstanceIndex(),
		CPUPercentage:  message.GetCpuPercentage(),
		MemBytes:       message.GetMemoryBytes(),
		DiskBytes:      message.GetDiskBytes(),
	}
}



// GetMapKeyFromAppData converts the combo of an app, space, and org into a hashmap key
func GetMapKeyFromAppData(orgName string, spaceName string, appName string) string {
	return fmt.Sprintf("%s/%s/%s", orgName, spaceName, appName)
}

func updateAppWithAppEvent(event Event)  {
	appName := event.AppName
	appOrg := event.OrgName
	appSpace := event.SpaceName

	appKey := GetMapKeyFromAppData(appOrg, appSpace, appName)
	appDetail := AppDetails[appKey]

	gcStatsMarker := "[GC"
	if strings.Contains(event.Msg, gcStatsMarker) {
		i, _ := strconv.ParseInt(event.SourceInstance, 10, 32)
		appDetail.Instances[i].GcStats = event.Msg
		logger.Println("Setting GC for app " + appKey)
	}

	AppDetails[appKey] = appDetail
	//logger.Println("Updated with App event " + appKey)

}

func updateAppWithContainerMetrics(event Event) {

	appName := event.AppName
	appOrg := event.OrgName
	appSpace := event.SpaceName

	appKey := GetMapKeyFromAppData(appOrg, appSpace, appName)

	var totalCPU float64 = 0
	var totalDiskUsage uint64 = 0
	var totalMemoryUsage uint64 = 0

	if 0 < len(AppDetails[appKey].Instances) {

		AppDetails[appKey].Instances[event.InstanceIndex].CellIP = event.CellIP
		AppDetails[appKey].Instances[event.InstanceIndex].CPUUsage = event.CPUPercentage
		AppDetails[appKey].Instances[event.InstanceIndex].MemoryUsage = event.MemBytes
		AppDetails[appKey].Instances[event.InstanceIndex].DiskUsage = event.DiskBytes

		for i := 0; i < len(AppDetails[appKey].Instances); i++ {
			totalCPU = totalCPU + event.CPUPercentage
			totalDiskUsage = totalDiskUsage + event.DiskBytes
			totalMemoryUsage = totalMemoryUsage + event.MemBytes
		}
	}

	AppDetails[appKey].EnvironmentSummary.TotalCPU = totalCPU
	AppDetails[appKey].EnvironmentSummary.TotalDiskUsage = totalDiskUsage
	AppDetails[appKey].EnvironmentSummary.TotalMemoryUsage = totalMemoryUsage

	//logger.Println("Updated with Container metrics " + appKey)
}

func updateAppDetails(event Event) {

	appName := event.AppName
	appOrg := event.OrgName
	appSpace := event.SpaceName

	appKey := GetMapKeyFromAppData(appOrg, appSpace, appName)
	appDetail := AppDetails[appKey]
	appDetail.Organization.Name = appOrg
	appDetail.Organization.ID = event.OrgID
	appDetail.Space.Name = appSpace
	appDetail.Space.ID = event.SpaceID
	appDetail.Name = appName
	appDetail.GUID = event.AppID

	appDetail.EventCount++
	appDetail.LastEventTime = time.Now().UnixNano()

	eventElapsed := time.Now().UnixNano() - appDetail.LastEventTime
	appDetail.ElapsedSinceLastEvent = eventElapsed / 1000000000
	totalElapsed := time.Now().UnixNano() - feedStarted
	elapsedSeconds := totalElapsed / 1000000000
	appDetail.RequestsPerSecond = float64(appDetail.EventCount) / float64(elapsedSeconds)
	appDetail.ElapsedSinceLastEvent = eventElapsed / 1000000000

	AppDetails[appKey] = appDetail
	//logger.Println("Updated with App Details " + appKey)

}

func getAppInfo(appGUID string) caching.App {
	if app := caching.GetAppInfo(appGUID); app.Name != "" {
		return app
	}
	caching.GetAppByGuid(appGUID)

	return caching.GetAppInfo(appGUID)
}

// LogMessage augments a raw message Envelope with log message metadata.
func LogMessage(msg *events.Envelope) Event {
	logMessage := msg.GetLogMessage()

	return Event{
		Origin:         msg.GetOrigin(),
		AppID:          logMessage.GetAppId(),
		Timestamp:      logMessage.GetTimestamp(),
		SourceType:     logMessage.GetSourceType(),
		SourceInstance: logMessage.GetSourceInstance(),
		MessageType:    logMessage.GetMessageType().String(),
		Msg:            string(logMessage.GetMessage()),
		Type:           msg.GetEventType().String(),
	}
}

// AnnotateWithAppData adds application specific details to an event by looking up the GUID in the cache.
func (e *Event) AnnotateWithAppData() {

	cfAppID := e.AppID
	appGUID := ""
	if cfAppID != "" {
		appGUID = fmt.Sprintf("%s", cfAppID)
	}

	if appGUID != "<nil>" && cfAppID != "" {
		appInfo := getAppInfo(appGUID)
		cfAppName := appInfo.Name
		cfSpaceID := appInfo.SpaceGuid
		cfSpaceName := appInfo.SpaceName
		cfOrgID := appInfo.OrgGuid
		cfOrgName := appInfo.OrgName

		if cfAppName != "" {
			e.AppName = cfAppName
		}

		if cfSpaceID != "" {
			e.SpaceID = cfSpaceID
		}

		if cfSpaceName != "" {
			e.SpaceName = cfSpaceName
		}

		if cfOrgID != "" {
			e.OrgID = cfOrgID
		}

		if cfOrgName != "" {
			e.OrgName = cfOrgName
		}
	}
}
