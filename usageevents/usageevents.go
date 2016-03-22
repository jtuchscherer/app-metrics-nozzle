package usageevents

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"

	"sync"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

// Event is a struct represented an event augmented/decorated with corresponding app/space/org data.
type Event struct {
	Fields logrus.Fields `json:"fields"`
	Msg    string        `json:"message"`
	Type   string        `json:"event_type"`
}

// ApplicationStat represents the observed metadata about an app, e.g. last router event time, etc.
type ApplicationStat struct {
	LastEventTime int64  `json:"last_event_time"`
	LastEvent     Event  `json:"last_event"`
	EventCount    int64  `json:"event_count"`
	AppName       string `json:"app_name"`
	OrgName       string `json:"org_name"`
	SpaceName     string `json:"space_name"`
}

type ApplicationDetail struct {
	Stats                 ApplicationStat `json:"stats"`
	RequestsPerSecond     float64         `json:"req_per_second"`
	ElapsedSinceLastEvent int64           `json:"elapsed_since_last_event"`
}

var mutex sync.Mutex

// AppStats is a map of app names to collected stats.
var AppStats = make(map[string]ApplicationStat)

var feedStarted int64

// ProcessEvents churns through the firehose channel, processing incoming events.
func ProcessEvents(in chan *events.Envelope) {
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
		if event.Fields["source_type"] == "RTR" {
			event.AnnotateWithAppData()
			updateAppStat(event)
		}
	}
	//fmt.Println("tick")
}

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

func updateAppStat(logEvent Event) {
	appName := logEvent.Fields["cf_app_name"].(string)
	appOrg := logEvent.Fields["cf_org_name"].(string)
	appSpace := logEvent.Fields["cf_space_name"].(string)

	appKey := GetMapKeyFromAppData(appOrg, appSpace, appName)
	appStat := AppStats[appKey]
	appStat.LastEventTime = time.Now().UnixNano()
	appStat.EventCount++
	appStat.AppName = appName
	appStat.SpaceName = appSpace
	appStat.OrgName = appOrg
	appStat.LastEvent = logEvent
	AppStats[appKey] = appStat
}

func getAppInfo(appGUID string) caching.App {
	if app := caching.GetAppInfo(appGUID); app.Name != "" {
		return app
	}
	caching.GetAppByGuid(appGUID)

	return caching.GetAppInfo(appGUID)
}

func LogMessage(msg *events.Envelope) Event {
	logMessage := msg.GetLogMessage()

	fields := logrus.Fields{
		"origin":          msg.GetOrigin(),
		"cf_app_id":       logMessage.GetAppId(),
		"timestamp":       logMessage.GetTimestamp(),
		"source_type":     logMessage.GetSourceType(),
		"message_type":    logMessage.GetMessageType().String(),
		"source_instance": logMessage.GetSourceInstance(),
	}

	return Event{
		Fields: fields,
		Msg:    string(logMessage.GetMessage()),
		Type:   msg.GetEventType().String(),
	}
}

func (e *Event) AnnotateWithAppData() {

	cf_app_id := e.Fields["cf_app_id"]
	appGuid := ""
	if cf_app_id != nil {
		appGuid = fmt.Sprintf("%s", cf_app_id)
	}

	if cf_app_id != nil && appGuid != "<nil>" && cf_app_id != "" {
		appInfo := getAppInfo(appGuid)
		cf_app_name := appInfo.Name
		cf_space_id := appInfo.SpaceGuid
		cf_space_name := appInfo.SpaceName
		cf_org_id := appInfo.OrgGuid
		cf_org_name := appInfo.OrgName

		if cf_app_name != "" {
			e.Fields["cf_app_name"] = cf_app_name
		}

		if cf_space_id != "" {
			e.Fields["cf_space_id"] = cf_space_id
		}

		if cf_space_name != "" {
			e.Fields["cf_space_name"] = cf_space_name
		}

		if cf_org_id != "" {
			e.Fields["cf_org_id"] = cf_org_id
		}

		if cf_org_name != "" {
			e.Fields["cf_org_name"] = cf_org_name
		}
	}
}

func (e *Event) AnnotateWithMetaData(extraFields map[string]string) {
	e.Fields["cf_origin"] = "firehose"
	e.Fields["event_type"] = e.Type
	for k, v := range extraFields {
		e.Fields[k] = v
	}
}
