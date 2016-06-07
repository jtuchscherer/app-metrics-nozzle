package usageevents

//http://json2struct.mervine.net/
type App struct {
	Buildpack            string `json:"build pack"`
	ElapsedSinceLastEvent int    `json:"elapsed_since_last_event"`
	Environment           struct {
				      EndpointURL string `json:"endpoint_url"`
				      Log         string `json:"log"`
			      } `json:"environment"`
	EnvironmentSummary    struct {
				      TotalCPU               string `json:"total_cpu"`
				      TotalDiskProvisioned   string `json:"total_disk_provisioned"`
				      TotalDiskUsage         string `json:"total_disk_usage"`
				      TotalMemoryProvisioned string `json:"total_memory_provisioned"`
				      TotalMemoryUsage       string `json:"total_memory_usage"`
			      } `json:"environment_summary"`
	EventCount            int    `json:"event_count"`
	GUID                  string `json:"guid"`
	InstanceCount         struct {
				      Configured int `json:"configured"`
				      Running    int `json:"running"`
			      } `json:"instance_count"`
	Instances             []struct {
		CPUUsage    string `json:"cpu_usage"`
		DiskQuota   string `json:"disk_quota"`
		DiskUsage   string `json:"disk_usage"`
		GcStats     struct {
				    Details string `json:"details"`
			    } `json:"gc_stats"`
		ID          int    `json:"id"`
		MemoryUsage string `json:"memory_usage"`
		State       string `json:"state"`
		Uptime      int    `json:"uptime"`
	} `json:"instances"`
	LastEvent             struct {
				      Message   string `json:"message"`
				      Timestamp int64    `json:"timestamp"`
			      } `json:"last_event"`
	Name                  string `json:"name"`
	Organization          struct {
				      ID   string `json:"id"`
				      Name string `json:"name"`
			      } `json:"organization"`
	RequestsPerSecond     int      `json:"requests_per_second"`
	Routes                []string `json:"routes"`
	Space                 struct {
				      ID   string `json:"id"`
				      Name string `json:"name"`
			      } `json:"space"`
	State                 string `json:"state"`
	SystemNumber          string `json:"system_number"`
}


