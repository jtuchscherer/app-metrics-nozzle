package usageevents

import (
	"testing"
	//"github.com/pivotalservices/_app-usage-nozzle/usageevents"
	 "fmt"
	//"encoding/json"
	//"time"
	"github.com/davecgh/go-spew/spew"
)



func TestReverse(t *testing.T) {

	appDetails := App{}

	appDetails.InstanceCount = doSomething("")

	spew.Dump(appDetails)
}

func doSomething(s string) InstanceCount{
	fmt.Println("doing something", s)
	return InstanceCount{Configured:3,Running:2}
}
