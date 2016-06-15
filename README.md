# App Usage Nozzle

This is a nozzle for the Cloud Foundry firehose component. It will ingest router events for every application it can detect and use the timestamps on those events to compute usage metrics. The first usage for this nozzle will be to determine if an application is unused by tracking the last time a request was routed to it.

## REST API
This application exposes a RESTful API that allows consumers to query application usage statistics. This API is described in the following table:

| Resource        | Method           | Description  |
| --- | --- | --- |
| `/api/apps` | GET | Queries the list of all _deployed_ applications. This is all applications in all organizations, in all spaces. The results will be a map whose key is a string of the format `[org]/[space]/[app name]`. |
| `/api/apps/[org]/[space]/[app]` | GET | Obtains application detail information, including time-based usage statistics as of the time of request, including elapsed time (in seconds) since the last event was received, and the requests per second for the app |

### JSON Payloads
This is a sample of what the JSON response looks like for `/api/apps`:

```javascript
"some-user-org/some-space/some-app": {
    "last_event_time": 0,
    "last_event": {
      "message": "",
      "event_type": "",
      "origin": "",
      "app_id": "",
      "timestamp": 0,
      "source_type": "",
      "message_type": "",
      "source_instance": "",
      "app_name": "",
      "org_name": "",
      "space_name": "",
      "org_id": "",
      "space_id": ""
    },
    "event_count": 0,
    "app_name": "CATS-persistent-app",
    "org_name": "CATS-persistent-org",
    "space_name": "CATS-persistent-space"
  },
"org/space/app" : {},
```

If the `last_event_time` field is `0` that indicates that no _router_ events for that application have been discovered _since the nozzle was started_.

The following is an example of application detail information, which includes computed metrics such as throughput (**req_per_second**) and elapsed time since last event, which can be used to determine the degree of application idleness.

```javascript
{
  "stats": {
    "last_event_time": 1458759091580510463,
    "last_event": {
      "message": "some.beverage.service.com - [23/03/2016:18:51:31 +0000] \"GET /beverages HTTP/1.1\" 200 0 122 \"-\" \"-\" 192.168.8.1:58827 x_forwarded_for:\"192.168.11.231\" x_forwarded_proto:\"http\" vcap_request_id:1cda921d-f528-41a2-724a-6b1aa1e4b350 response_time:0.004863968 app_id:e35995c2-b7d9-4e0f-800a-bfc312446dd4\n",
      "event_type": "LogMessage",
      "origin": "router__0",
      "app_id": "e35995c2-b7d9-4e0f-800a-bfc312446dd4",
      "timestamp": 1458759091584918617,
      "source_type": "RTR",
      "message_type": "OUT",
      "source_instance": "0",
      "app_name": "some-beverage-service",
      "org_name": "some-org",
      "space_name": "some-space",
      "org_id": "dc996683-574b-4d61-8bde-bce8065ae044",
      "space_id": "5e0f626e-9fdd-4e86-abe0-79232f9263b3"
    },
    "event_count": 136,
    "app_name": "some-beverage-service",
    "org_name": "some-org",
    "space_name": "some-space"
  },
  "req_per_second": 0.032037691401649,
  "elapsed_since_last_event": 21
}
```

## Installation
To install this application, it should be run as an app within Cloud Foundry. So, the first thing you'll need to do is push the app. There is a `manifest.yml` already included in the project, so you can just do:

```
cf push app-usage-nozzle --no-start
```

The `no-start` is important because we have not yet defined the environment variables that allow the application to connect to the Firehose and begin monitoring router requests. We want to end up with a set of environment variables that looks like this when we issue a `cf env app-metrics-nozzle` command:

```
User-Provided:
API_ENDPOINT: https://api.run.pez.pivotal.io
CF_PULL_TIME: 9999s
FIREHOSE_PASSWORD: (this is a secret)
FIREHOSE_SUBSCRIPTION_ID: app-usage-nozzle
FIREHOSE_USER: (this is also secret)
SKIP_SSL_VALIDATION: true
```
Once you've set these environment variables with `cf set-env (app) (var) (value)` you can just start the application usage nozzle via `cf start`. Make sure the application has come up by hitting the API endpoint. Depending on how large of a foundation in which it was deployed, it can take _several minutes_ for the cache of application metadata to fill up.
