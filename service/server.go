/*
Copyright 2016 Pivotal

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

type userProvider struct {
	username string
	password string
}

func (u *userProvider) credsMatch(username, password string) bool {
	return username == u.username && password == u.password
}

// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {
	up := userProvider{
		username: os.Getenv("USERNAME"),
		password: os.Getenv("PASSWORD"),
	}

	formatter := render.New(render.Options{
		IndentJSON: true,
	})

	n := negroni.Classic()
	mx := mux.NewRouter()

	initRoutes(mx, formatter, up)

	n.UseHandler(mx)
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render, up userProvider) {
	mx.HandleFunc("/api/apps/{org}/{space}/{app}/{instance_id}", authenticate(appInstanceHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/apps/{org}/{space}/{app}", authenticate(appHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/apps/{org}/{space}", authenticate(appSpaceHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/apps/{org}", authenticate(appOrgHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/apps", authenticate(appAllHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/orgs/{org}", authenticate(orgDetailsHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/orgs", authenticate(orgsHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/orgs/{org}/users", authenticate(orgsUsersHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/orgs/{org}/{role}/users", authenticate(orgsUsersByRoleHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/spaces/{space}", authenticate(spaceDetailsHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/spaces/{space}/users", authenticate(spacesUsersHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/spaces/{space}/{role}/users", authenticate(spacesUsersByRoleHandler(formatter), up)).Methods("GET")
	mx.HandleFunc("/api/spaces", authenticate(spaceHandler(formatter), up)).Methods("GET")

}

func authenticate(h http.HandlerFunc, userProvider userProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		doAuthentication(w, r, h, userProvider)
	}
}

func doAuthentication(w http.ResponseWriter, r *http.Request, innerHandler func(w http.ResponseWriter, r *http.Request), userProvider userProvider) {
	w.Header().Set("WWW-Authenticate", `Basic realm="pprof"`)

	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 || s[0] != "Basic" {
		http.Error(w, "Invalid authorization header", 401)
		return
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}

	credentials := strings.SplitN(string(b), ":", 2)
	if len(credentials) != 2 {
		http.Error(w, "Invalid authorization header", 401)
		return
	}

	if !userProvider.credsMatch(credentials[0], credentials[1]) {
		http.Error(w, "Not authorized", 401)
		return
	}

	innerHandler(w, r)
}
