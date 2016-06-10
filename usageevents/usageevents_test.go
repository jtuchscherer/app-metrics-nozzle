package usageevents

import (
	"testing"
	//"github.com/pivotalservices/_app-usage-nozzle/usageevents"
	 "fmt"
	//"encoding/json"
	//"time"
	//"github.com/davecgh/go-spew/spew"
	//"github.com/pivotalservices/app-usage-nozzle/usageevents"
	"strings"
)



func TestReverse(t *testing.T) {
	appDetail := App{GUID:"asdfadsfa", Name:"my app"}
	appDetail.Routes = make([]string, 3)
	appDetail.Routes[0] = "zero"
	appDetail.Routes[1] = "one"
	appDetail.Routes[2] = "two"
	//fmt.Println(appDetail.Routes)

	msg := "2016-06-09T14:35:58.087+0000: [GC (Allocation Failure) [PSYoungGen: 229368K->32759K(229376K)] 247484K->58983K(753664K), 0.0168215 secs] [Times: user=0.05 sys=0.01, real=0.01 secs] "

	gcStatsMarker := "[GC"
	fmt.Println(strings.Contains(msg, gcStatsMarker))

}

func doSomething(s string) InstanceCount{
	fmt.Println("doing something", s)
	return InstanceCount{Configured:3,Running:2}
}
