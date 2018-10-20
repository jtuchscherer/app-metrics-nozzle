# App Metrics Nozzle

This is a nozzle for the Cloud Foundry firehose component. It will ingest router events for every application it can detect and use the timestamps on those events to compute usage metrics. The first usage for this nozzle will be to determine if an application is unused by tracking the last time a request was routed to it.
It will also make REST api calls to Cloud Controller and cache application, and user data for the interval specified in configuration.

## REST API
This application exposes a RESTful API that allows consumers to query application usage statistics. This API is described in the following table:

| Resource        | Method           | Description  |
| --- | --- | --- |
| `/api/apps` | GET | Queries the list of all _deployed_ applications. This is all applications in all organizations, in all spaces. The results will be a map whose key is a string of the format `[org]/[space]/[app name]`. |
| `/api/apps/[org]/[space]/[app]` | GET | Obtains application detail information, including time-based usage statistics as of the time of request, including elapsed time (in seconds) since the last event was received, and the requests per second for the app. |
| `/api/apps/[org]/[space]/[app]/[instance_id]` | GET | Obtains information for the instance of the application including IP, CPI usage and Memory usage. |
| `/api/apps/[org]/[space]` | GET | Obtains application details deployed in specified space. |
| `/api/apps/[org]` | GET | Obtains application details deployed in specified organization. |
| `/api/orgs` | GET | Obtains names and guids of all organizations. |
| `/api/orgs/[org]` | GET | Obtains name and guid of an organization. |
| `/api/orgs/[org]/users` | GET | Returns information about users of specified organization. |
| `/api/orgs/[org]/[role]/users` | GET | Returns users of specified organization by role. |
| `/api/spaces` | GET | Returns a list of spaces. |
| `/api/spaces/[space]` | GET | Returns space details. |
| `/api/spaces/[space]/users` | GET | Returns users of specified space. |
| `/api/spaces/[space]/[role]/users` | GET | Returns users of specified space by role. |

### JSON Payloads
This is a sample of what the JSON response looks like for the app `/api/apps`:

```javascript
{
  "johannes-org/development/app-metrics-nozzle": {
    "buildpack": "https://github.com/cloudfoundry/go-buildpack.git",
    "diego": true,
    "environment": {
      "API_ENDPOINT": "https://api.run.pivotal.io",
      "CF_PULL_TIME": "300s",
      "DOPPLER_ENDPOINT": "wss://doppler.run.pivotal.io:443",
      "FIREHOSE_PASSWORD": "[[PRIVATE]]",
      "FIREHOSE_USER": "[[PRIVATE]]",
      "GOPACKAGENAME": "app-metrics-nozzle",
      "PASSWORD": "[[PRIVATE]]",
      "SKIP_SSL_VALIDATION": "false",
      "USERNAME": "[[PRIVATE]]"
    },
    "environment_summary": {
      "total_cpu": 1.2153818869424569,
      "total_disk_configured": 268435456,
      "total_disk_provisioned": 268435456,
      "total_disk_usage": 51855360,
      "total_memory_configured": 268435456,
      "total_memory_provisioned": 268435456,
      "total_memory_usage": 51267356
    },
    "guid": "dcfbab8d-46cb-475d-a95b-da2e67ee3312",
    "instance_count": {
      "configured": 1,
      "running": 1
    },
    "instances": Array[1][
      {
        "cell_ip": "10.10.148.165",
        "cpu_usage": 1.2153818869424569,
        "disk_usage": 51855360,
        "gc_stats": "",
        "index": 0,
        "memory_usage": 51267356,
        "uptime": 45049,
        "since": 1539962977,
        "state": "RUNNING",
        "last_event": "2018-10-20 04:01:24.970336982 +0000 UTC"
      }
    ],
    "name": "app-metrics-nozzle",
    "organization": {
      "id": "[[PRIVATE]]",
      "name": "[[PRIVATE]]"
    },
    "event_count": 1,
    "last_event_time": 1540008031613128991,
    "requests_per_second": 0.000022214817283127847,
    "elapsed_since_last_event": 0,
    "routes": Array[1][
      "app-metrics-nozzle.cfapps.io"
    ],
    "space": {
      "id": "[[PRIVATE]]5",
      "name": "[[PRIVATE]]"
    },
    "state": "STARTED",
    "fetch_time": "2018-10-20 04:00:27.673069672 +0000 UTC"
  },
  [...]
```

If the `last_event_time` field is `0` that indicates that no _router_ events for that application have been discovered _since the nozzle was started_.

## Installation
Run glide install to pull dependencies into vendor directory.
To install this application, it should be run as an app within Cloud Foundry. So, the first thing you'll need to do is push the app. There is a `manifest.yml` already included in the project. After you changed the environment variables in the manifest.yml, you can just run the following command:

```
cf push app-metrics-nozzle
```

DOPPLER_ENDPOINT can be obtained by running
```bash
cat ~/.cf/config.json | grep Doppler
```

