package usageevents

import (
	"testing"
	//"github.com/pivotalservices/_app-usage-nozzle/usageevents"
	 "fmt"
	//"encoding/json"
	"time"
)



func TestReverse(t *testing.T) {

	for {
		time.Sleep(7 * time.Second)
		go doSomething("from polling 1")
	}

	//key := ApplicationStat{AppName: "app name", SpaceName: "app space", OrgName: "org"}
	//
	//b, err := json.Marshal(key)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(string(b))
}

func doSomething(s string) {
	fmt.Println("doing something", s)
}
