package usageevents

import (
	"testing"
	//"github.com/pivotalservices/_app-usage-nozzle/usageevents"
	 "fmt"
	//"encoding/json"
	//"time"
	//"github.com/davecgh/go-spew/spew"
	//"github.com/pivotalservices/app-usage-nozzle/usageevents"
	//"strings"
	"strconv"
	"os"
	"log"
)



func TestReverse(t *testing.T) {

	var instance int32 = 1
	var totalCPU float64 = 0.2
	var totalDiskUsage uint64 = 1
	var totalMemoryUsage uint64 = 1

	totalCPU = totalCPU + .3
	totalDiskUsage = totalDiskUsage + 1
	totalMemoryUsage = totalMemoryUsage + 1

	logMsg := fmt.Sprintf("LOG GC messge instance %s %s--%s--%s", fmt.Sprintf("%v",instance),
		strconv.FormatFloat(totalCPU, 'f', 6, 64),
		strconv.FormatUint(totalDiskUsage, 10),
		strconv.FormatUint(totalMemoryUsage, 10))

	logger := log.New(os.Stdout, "", 0)
	logger.Println(logMsg)

}

func doSomething(s string) InstanceCount{
	fmt.Println("doing something", s)
	return InstanceCount{Configured:3,Running:2}
}
