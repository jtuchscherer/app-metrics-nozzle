package api

import (
	"testing"
	//"github.com/pivotalservices/_app-usage-nozzle/usageevents"
	// "fmt"
	//"encoding/json"
	//"time"
	//"github.com/davecgh/go-spew/spew"
	//"github.com/pivotalservices/app-usage-nozzle/usageevents"
	"app-usage-nozzle/domain"
	"fmt"
	"time"
)


var AppDetails = make(map[string]domain.App)

func TestReverse(t *testing.T) {

	now := time.Now()

	fmt.Println("now:", now)

	then := now.Add(-10 * time.Minute)
	fmt.Println("10 minutes ago:", then)
}
