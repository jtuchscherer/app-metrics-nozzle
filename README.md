# App Usage Nozzle

This is a nozzle for the Cloud Foundry firehose component. It will ingest router events for every application it can detect and use the timestamps on those events to compute usage metrics. The first usage for this nozzle will be to determine if an application is unused by tracking the last time a request was routed to it.
