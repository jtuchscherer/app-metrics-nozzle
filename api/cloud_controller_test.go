package api

import (
	"testing"
	//"github.com/pivotalservices/_app-usage-nozzle/usageevents"
	// "fmt"
	//"encoding/json"
	//"time"
	//"github.com/davecgh/go-spew/spew"
	//"github.com/pivotalservices/app-usage-nozzle/usageevents"
	"github.com/davecgh/go-spew/spew"
	"app-usage-nozzle/domain"
)


var AppDetails = make(map[string]domain.App)

func TestReverse(t *testing.T) {

	appKey := "bb7b3c89-0a7f-47f7-9dd3-5e4fbd8ded6c"

	//app := usageevents.AppDetails[appKey]
	//

	app := domain.App{GUID:appKey}

	AnnotateWithCloudControllerData(&app)

	spew.Dump(app)
}
