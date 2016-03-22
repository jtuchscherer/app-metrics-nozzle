package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CrowdSurge/banner"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/boltdb/bolt"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/firehose"
	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pivotalservices/app-usage-nozzle/service"
	"github.com/pivotalservices/app-usage-nozzle/usageevents"
)

var (
	debug             = kingpin.Flag("debug", "Enable debug mode. This disables forwarding to syslog").Default("false").OverrideDefaultFromEnvar("DEBUG").Bool()
	apiEndpoint       = kingpin.Flag("api-endpoint", "Api endpoint address. For bosh-lite installation of CF: https://api.10.244.0.34.xip.io").OverrideDefaultFromEnvar("API_ENDPOINT").Required().String()
	dopplerEndpoint   = kingpin.Flag("doppler-endpoint", "Overwrite default doppler endpoint return by /v2/info").OverrideDefaultFromEnvar("DOPPLER_ENDPOINT").String()
	subscriptionID    = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").String()
	user              = kingpin.Flag("user", "Admin user.").Default("admin").OverrideDefaultFromEnvar("FIREHOSE_USER").String()
	password          = kingpin.Flag("password", "Admin password.").Default("admin").OverrideDefaultFromEnvar("FIREHOSE_PASSWORD").String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Bool()
	boltDatabasePath  = kingpin.Flag("boltdb-path", "Bolt Database path ").Default("my.db").OverrideDefaultFromEnvar("BOLTDB_PATH").String()
	tickerTime        = kingpin.Flag("cc-pull-time", "CloudController Polling time in sec").Default("60s").OverrideDefaultFromEnvar("CF_PULL_TIME").Duration()
)

const (
	version = "0.0.1"
)

func main() {

	banner.Print("usage nozzle")
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	// Start web server
	go func() {
		server := service.NewServer()
		server.Run(":" + port)
	}()

	kingpin.Version(version)
	kingpin.Parse()

	log.Println(fmt.Sprintf("Starting app-usage-nozzle %s ", version))

	c := cfclient.Config{
		ApiAddress:        *apiEndpoint,
		Username:          *user,
		Password:          *password,
		SkipSslValidation: *skipSSLValidation,
	}
	cfClient := cfclient.NewClient(&c)

	if len(*dopplerEndpoint) > 0 {
		cfClient.Endpoint.DopplerEndpoint = *dopplerEndpoint
	}
	log.Println(fmt.Sprintf("Using %s as doppler endpoint", cfClient.Endpoint.DopplerEndpoint))

	//Use bolt for in-memory  - file caching
	db, err := bolt.Open(*boltDatabasePath, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		log.Fatal("Error opening bolt db: ", err)
		os.Exit(1)

	}
	defer db.Close()

	caching.SetCfClient(cfClient)
	caching.SetAppDb(db)
	caching.CreateBucket()

	//Let's Update the database the first time
	log.Println("Start filling app/space/org cache.")
	apps := caching.GetAllApp()
	for idx := range apps {
		org := apps[idx].OrgName
		space := apps[idx].SpaceName
		app := apps[idx].Name
		key := usageevents.GetMapKeyFromAppData(org, space, app)
		usageevents.AppStats[key] = usageevents.ApplicationStat{AppName: app, SpaceName: space, OrgName: org}
	}

	log.Println(fmt.Sprintf("Done filling cache! Found [%d] Apps", len(apps)))

	// Ticker Pooling the CC every X sec
	ccPolling := time.NewTicker(*tickerTime)

	go func() {
		for range ccPolling.C {
			log.Println("Re-loading application cache.")
			apps = caching.GetAllApp()
		}
	}()

	firehose := firehose.CreateFirehoseChan(cfClient.Endpoint.DopplerEndpoint, cfClient.GetToken(), *subscriptionID, *skipSSLValidation)
	if firehose != nil {
		log.Println("Firehose Subscription Succesfull! Routing events...")
		usageevents.ProcessEvents(firehose)
	} else {
		log.Fatal("Failed connecting to Firehose...Please check settings and try again!")
	}
}
