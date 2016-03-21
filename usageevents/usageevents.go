package usageevents

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/utils"

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
	EventCount    uint64 `json:"event_count"`
	AppName       string `json:"app_name"`
	OrgName       string `json:"org_name"`
	SpaceName     string `json:"space_name"`
}

var mutex sync.Mutex

// AppStats is a map of app names to collected stats.
var AppStats = make(map[string]ApplicationStat)

// ProcessEvents churns through the firehose channel, processing incoming events.
func ProcessEvents(in chan *events.Envelope) {
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
	if eventType == events.Envelope_CounterEvent {
		event = CounterEvent(msg)
	}
	fmt.Println("tick")
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

func HttpStart(msg *events.Envelope) Event {
	httpStart := msg.GetHttpStart()

	fields := logrus.Fields{
		"origin":            msg.GetOrigin(),
		"cf_app_id":         utils.FormatUUID(httpStart.GetApplicationId()),
		"instance_id":       httpStart.GetInstanceId(),
		"instance_index":    httpStart.GetInstanceIndex(),
		"method":            httpStart.GetMethod(),
		"parent_request_id": utils.FormatUUID(httpStart.GetParentRequestId()),
		"peer_type":         httpStart.GetPeerType(),
		"request_id":        utils.FormatUUID(httpStart.GetRequestId()),
		"remote_addr":       httpStart.GetRemoteAddress(),
		"timestamp":         httpStart.GetTimestamp(),
		"uri":               httpStart.GetUri(),
		"user_agent":        httpStart.GetUserAgent(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func HttpStop(msg *events.Envelope) Event {
	httpStop := msg.GetHttpStop()

	fields := logrus.Fields{
		"origin":         msg.GetOrigin(),
		"cf_app_id":      utils.FormatUUID(httpStop.GetApplicationId()),
		"content_length": httpStop.GetContentLength(),
		"peer_type":      httpStop.GetPeerType(),
		"request_id":     utils.FormatUUID(httpStop.GetRequestId()),
		"status_code":    httpStop.GetStatusCode(),
		"timestamp":      httpStop.GetTimestamp(),
		"uri":            httpStop.GetUri(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func HttpStartStop(msg *events.Envelope) Event {
	httpStartStop := msg.GetHttpStartStop()

	fields := logrus.Fields{
		"origin":         msg.GetOrigin(),
		"cf_app_id":      utils.FormatUUID(httpStartStop.GetApplicationId()),
		"content_length": httpStartStop.GetContentLength(),
		"instance_id":    httpStartStop.GetInstanceId(),
		"instance_index": httpStartStop.GetInstanceIndex(),
		"method":         httpStartStop.GetMethod(),
		//	"parent_request_id": utils.FormatUUID(httpStartStop.GetParentRequestId()),
		"peer_type":       httpStartStop.GetPeerType(),
		"remote_addr":     httpStartStop.GetRemoteAddress(),
		"request_id":      utils.FormatUUID(httpStartStop.GetRequestId()),
		"start_timestamp": httpStartStop.GetStartTimestamp(),
		"status_code":     httpStartStop.GetStatusCode(),
		"stop_timestamp":  httpStartStop.GetStopTimestamp(),
		"uri":             httpStartStop.GetUri(),
		"user_agent":      httpStartStop.GetUserAgent(),
		"duration_ms":     (((httpStartStop.GetStopTimestamp() - httpStartStop.GetStartTimestamp()) / 1000) / 1000),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
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

func ValueMetric(msg *events.Envelope) Event {
	valMetric := msg.GetValueMetric()

	fields := logrus.Fields{
		"origin": msg.GetOrigin(),
		"name":   valMetric.GetName(),
		"unit":   valMetric.GetUnit(),
		"value":  valMetric.GetValue(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func CounterEvent(msg *events.Envelope) Event {
	counterEvent := msg.GetCounterEvent()

	fields := logrus.Fields{
		"origin": msg.GetOrigin(),
		"name":   counterEvent.GetName(),
		"delta":  counterEvent.GetDelta(),
		"total":  counterEvent.GetTotal(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func ErrorEvent(msg *events.Envelope) Event {
	errorEvent := msg.GetError()

	fields := logrus.Fields{
		"origin": msg.GetOrigin(),
		"code":   errorEvent.GetCode(),
		"delta":  errorEvent.GetSource(),
	}

	return Event{
		Fields: fields,
		Msg:    errorEvent.GetMessage(),
		Type:   msg.GetEventType().String(),
	}
}

func ContainerMetric(msg *events.Envelope) Event {
	containerMetric := msg.GetContainerMetric()

	fields := logrus.Fields{
		"origin":         msg.GetOrigin(),
		"cf_app_id":      containerMetric.GetApplicationId(),
		"cpu_percentage": containerMetric.GetCpuPercentage(),
		"disk_bytes":     containerMetric.GetDiskBytes(),
		"instance_index": containerMetric.GetInstanceIndex(),
		"memory_bytes":   containerMetric.GetMemoryBytes(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
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
