package service

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pivotalservices/app-usage-nozzle/usageevents"
	"github.com/unrolled/render"
)

func appCollectionHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, usageevents.AppStats)
	}
}

func singleAppHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		app := vars["app"]
		org := vars["org"]
		space := vars["space"]
		key := usageevents.GetMapKeyFromAppData(org, space, app)

		fmt.Printf("Retrieving app at key : %s\n", key)
		stat, exists := usageevents.AppStats[key]
		if exists {
			formatter.JSON(w, http.StatusOK, stat)
		} else {
			formatter.JSON(w, http.StatusNotFound, "No such app")
		}
	}
}
