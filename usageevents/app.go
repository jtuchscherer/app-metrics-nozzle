package usageevents

//import "github.com/cloudfoundry-community/firehose-to-syslog/Godeps/_workspace/src/github.com/cloudfoundry/sonde-go/events"

//import "github.com/pquerna/ffjson/shared"

//http://json2struct.mervine.net/
type InstanceCount  struct {
	Configured int `json:"configured"`
	Running    int `json:"running"`
}

type Instances struct {
	CellIP        string `json:"cell_ip"`
	CPUUsage      float64 `json:"cpu_usage"`   //CPUPercentage
	DiskUsage     uint64 `json:"disk_usage"`   //DiskBytes
	GcStats       string `json:"gc_stats"`
	InstanceIndex int64    `json:"index"`
	MemoryUsage   uint64 `json:"memory_usage"` //MemBytes
	Uptime        int32 `json:"uptime"`        //todo calculate
	Since         int32 `json:"since"`
	State         string `json:"state"`
}
type App struct {
	Buildpack             string `json:"buildpack"`
	Diego                 bool `json:"diego"`

	Environment           map[string]interface{} `json:"environment"`
	EnvironmentSummary    struct {
				      TotalCPU               string `json:"total_cpu"`          //todo calculate from instances add these from instances
				      TotalDiskConfigured   int32 `json:"total_disk_configured"`
				      TotalDiskProvisioned   int32 `json:"total_disk_provisioned"`
				      TotalDiskUsage         string `json:"total_disk_usage"`   //todo calculate from instances add these from instances
				      TotalMemoryConfigured int32 `json:"total_memory_congigured"`
				      TotalMemoryProvisioned int32 `json:"total_memory_provisioned"`
				      TotalMemoryUsage       string `json:"total_memory_usage"` //todo calculate from instances add these from instances
			      } `json:"environment_summary"`
	GUID                  string `json:"guid"`
	InstanceCount          `json:"instance_count"`
	Instances             []Instances  `json:"instances"`
	Name                  string `json:"name"`
	Organization          struct {
				      ID   string `json:"id"`
				      Name string `json:"name"`
			      } `json:"organization"`
	EventCount            int64 `json:"event_count"`
	LastEventTime         int64   `json:"last_event_time"`
	RequestsPerSecond     float64      `json:"requests_per_second"`
	ElapsedSinceLastEvent int64    `json:"elapsed_since_last_event"`
	Routes                []string `json:"routes"`
	Space                 struct {
				      ID   string `json:"id"`
				      Name string `json:"name"`
			      } `json:"space"`
	State                 string `json:"state"`
	SystemNumber          string `json:"system_number"`
}


