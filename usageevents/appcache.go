package usageevents

import "github.com/cloudfoundry-community/firehose-to-syslog/caching"



type AppCache struct{
}

type CachedApp interface {
	GetAppByGuid(appGuid string) []caching.App
	GetAppInfo(appGuid string) caching.App
	GetAllApp() []caching.App
}

func (c *AppCache)GetAppByGuid(appGuid string) []caching.App{
	return caching.GetAppByGuid(appGuid)
}

func (c *AppCache)GetAppInfo(appGuid string) caching.App{
	return caching.GetAppInfo(appGuid)
}

func (c *AppCache)GetAllApp() []caching.App {
	return caching.GetAllApp()
}
