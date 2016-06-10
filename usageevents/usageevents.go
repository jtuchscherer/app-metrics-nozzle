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
	"log"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"sync"
	"time"
	"github.com/cloudfoundry/sonde-go/events"
	cfclient "github.com/jtgammon/go-cfclient"
	"os"
	"strings"
	"strconv"
	//"github.com/davecgh/go-spew/spew"
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
	CellIP	       string `json:"cell_ip"`
	InstanceIndex  int32  `json:"instance_index"`
	CPUPercentage  float64 `json:"cpu_percentage"`
	MemBytes       uint64 `json:"mem_bytes"`
	DiskBytes      uint64 `json:"disk_bytes"`
}

// ApplicationStat represents the observed metadata about an app, e.g. last router event time, etc.
type ApplicationStat struct {
	LastEventTime int64   `json:"last_event_time"`
	LastEvent     Event   `json:"last_event"`
	EventCount    int64   `json:"event_count"`
	AppName       string  `json:"app_name"`
	OrgName       string  `json:"org_name"`
	SpaceName     string  `json:"space_name"`
	LastEventRPS  float64 `json:"last_event_rps"`
}

// ApplicationDetail represents a time snapshot of the RPS and elapsed time since last event for an app
type ApplicationDetail struct {
	Stats                 ApplicationStat `json:"stats"`
	RequestsPerSecond     float64         `json:"req_per_second"`
	ElapsedSinceLastEvent int64           `json:"elapsed_since_last_event"`
}

var mutex sync.Mutex

// AppStats is a map of app names to collected stats.
var AppStats = make(map[string]ApplicationStat)

var AppDetails = make(map[string]App)

var feedStarted int64

// ProcessEvents churns through the firehose channel, processing incoming events.
func ProcessEvents(in chan *events.Envelope) {
	feedStarted = time.Now().UnixNano()
	for msg := range in {
		processEvent(msg)
	}
}

func UpdateCloudContollerData(appId string) {

	logger := log.New(os.Stdout, "", 0)
	logger.Println("Re-loading application data from Cloud Controller." + appId)

	c := cfclient.Config{
		ApiAddress:        "https://api.run.haas-41.pez.pivotal.io",
		Username:          "admin",
		Password:          "cb0a40f8d6360eaed442",
		SkipSslValidation: true,
	}
	client, _ := cfclient.NewClient(&c)

	app, _ := client.AppByGuid(appId)

	instances, _ := client.GetAppInstances(app.Guid)
	runnintCount := 0
	instanceUp := "RUNNING"

	space, _ := app.Space()
	org, _ := space.Org()

	appKey := GetMapKeyFromAppData(org.Name, space.Name, app.Name)

	logger.Println("-->" + appKey + "---" + strconv.Itoa(runnintCount) + "/" + strconv.Itoa(len(instances)))

	appDetail := AppDetails[appKey]
	appDetail.Diego = app.Diego
	appDetail.Buildpack = app.Buildpack
	appDetail.Instances = make([]Instances, int64(len(instances)))

	for idx, eachInstance := range instances {
		if strings.Compare(instanceUp, eachInstance.State) == 0 {
			runnintCount++;
		}
		i, _ := strconv.ParseInt(idx, 10, 32)
		appDetail.Instances[i].InstanceIndex = i
		appDetail.Instances[i].State = eachInstance.State
		appDetail.Instances[i].Since = eachInstance.Since
		appDetail.Instances[i].Uptime = eachInstance.Uptime
	}

	if len(appDetail.Buildpack) == 0 { appDetail.Buildpack = app.DetectedBP }

	appDetail.Environment = app.Environment

	appDetail.Organization.ID = org.Guid
	appDetail.Organization.Name = org.Name

	appDetail.Space.ID = space.Guid
	appDetail.Space.Name = space.Name

	appDetail.InstanceCount.Configured = len(instances)
	appDetail.InstanceCount.Running = runnintCount

	appDetail.EnvironmentSummary.TotalDiskConfigured = app.DiskQuota
	appDetail.EnvironmentSummary.TotalMemoryConfigured = app.MemQuota

	appDetail.EnvironmentSummary.TotalDiskProvisioned = app.DiskQuota * int32(len(instances))
	appDetail.EnvironmentSummary.TotalMemoryProvisioned = app.MemQuota * int32(len(instances))

	if 0 < len(app.RouteData) {
		appDetail.Routes = make([]string, len(app.RouteData))
		for i := 0; i < len(app.RouteData); i++ {
			appDetail.Routes[i] = app.RouteData[i].Entity.Host + "." + app.RouteData[i].Entity.DomainData.Entity.Name
		}
	}

	appDetail.State = app.State


	AppDetails[appKey] = appDetail
}

func processEvent(msg *events.Envelope) {
	eventType := msg.GetEventType()

	var event Event
	if eventType == events.Envelope_LogMessage {
		event = LogMessage(msg)
		if event.SourceType == "RTR" {
			event.AnnotateWithAppData()
			//logger := log.New(os.Stdout, "", 0)
			//logger.Println("-------> Log message to parse " + event.Msg + " org " + event.OrgName + " space " + event.SpaceName)
			updateAppStat(event)
		}

		if event.SourceType == "APP" {
			event.AnnotateWithAppData()
			updateAppDetails(event)
		}
	}

	if eventType == events.Envelope_ContainerMetric {
		event = ContainerMetric(msg)
		event.AnnotateWithAppData()
		updateAppDetails(event)
	}
}

func ContainerMetric(msg *events.Envelope) Event {
	message := msg.GetContainerMetric()

	return Event{
		Origin:         msg.GetOrigin(),
		Type:           msg.GetEventType().String(),
		AppID:          message.GetApplicationId(),
		CellIP:		*msg.Ip,
		InstanceIndex:  message.GetInstanceIndex(),
		CPUPercentage:  message.GetCpuPercentage(),
		MemBytes:       message.GetMemoryBytes(),
		DiskBytes:      message.GetDiskBytes(),
	}
}
// CalculateDetailedStat takes application stats, uses the clock time, and calculates elapsed times and requests/second.
func CalculateDetailedStat(stat ApplicationStat) (detail ApplicationDetail) {
	detail.Stats = stat
	if len(stat.LastEvent.Type) > 0 {
		eventElapsed := time.Now().UnixNano() - stat.LastEventTime
		detail.ElapsedSinceLastEvent = eventElapsed / 1000000000
		totalElapsed := time.Now().UnixNano() - feedStarted
		elapsedSeconds := totalElapsed / 1000000000
		detail.RequestsPerSecond = float64(stat.EventCount) / float64(elapsedSeconds)
	}
	return
}

// GetMapKeyFromAppData converts the combo of an app, space, and org into a hashmap key
func GetMapKeyFromAppData(orgName string, spaceName string, appName string) string {
	return fmt.Sprintf("%s/%s/%s", orgName, spaceName, appName)
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

	var totalCPU float64 = 0
	var totalDiskUsage uint64 = 0
	var totalMemoryUsage uint64 = 0

	if 0 < len(appDetail.Instances) {

		appDetail.Instances[event.InstanceIndex].CellIP = event.CellIP
		appDetail.Instances[event.InstanceIndex].CPUUsage = event.CPUPercentage
		appDetail.Instances[event.InstanceIndex].MemoryUsage = event.MemBytes
		appDetail.Instances[event.InstanceIndex].DiskUsage = event.DiskBytes

		for i := 0; i < len(appDetail.Instances); i++ {
			totalCPU = totalCPU + event.CPUPercentage
			totalDiskUsage = totalDiskUsage + event.DiskBytes
			totalMemoryUsage = totalMemoryUsage + event.MemBytes
		}
	}

	appDetail.EnvironmentSummary.TotalCPU = totalCPU
	appDetail.EnvironmentSummary.TotalDiskUsage = totalDiskUsage
	appDetail.EnvironmentSummary.TotalMemoryUsage = totalMemoryUsage

	gcStatsMarker := "[GC"
	if strings.Contains(event.Msg, gcStatsMarker){
		i, _ := strconv.ParseInt(event.SourceInstance, 10, 32)

		appDetail.Instances[i].GcStats = event.Msg
	}

	appStats := AppStats[appKey]
	appDetail.RequestsPerSecond = appStats.LastEventRPS
	appDetail.EventCount = appStats.EventCount
	appDetail.LastEventTime = appStats.LastEventTime

	eventElapsed := time.Now().UnixNano() - appStats.LastEventTime
	appDetail.ElapsedSinceLastEvent = eventElapsed / 1000000000

	AppDetails[appKey] = appDetail
}

func updateAppStat(logEvent Event) {
	appName := logEvent.AppName
	appOrg := logEvent.OrgName
	appSpace := logEvent.SpaceName

	appKey := GetMapKeyFromAppData(appOrg, appSpace, appName)
	appStat := AppStats[appKey]
	appStat.LastEventTime = time.Now().UnixNano()
	appStat.EventCount++
	appStat.AppName = appName
	appStat.SpaceName = appSpace
	appStat.OrgName = appOrg
	appStat.LastEvent = logEvent

	detail := CalculateDetailedStat(appStat)
	appStat.LastEventRPS = detail.RequestsPerSecond
	AppStats[appKey] = appStat
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

func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r) - 1; i < len(r) / 2; i, j = i + 1, j - 1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
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
