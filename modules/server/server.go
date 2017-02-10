// Copyright 2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package server

import (
	"encoding/json"
	"net/http"

	"github.com/awslabs/ecs-secrets/modules/api"

	"github.com/awslabs/ecs-secrets/modules/store"
	"github.com/awslabs/ecs-secrets/modules/version"

	log "github.com/cihub/seelog"
	"github.com/gorilla/mux"
)

const listeningPort = "8080"

// Server interface defines methods to route and serve requests when running
// in 'daemon' mode
type Server interface {
	Router() *mux.Router
	Serve() error
}

type server struct {
	secretStore store.Store
}

type versionResponse struct {
	application        string `json:"application"`
	applicationVersion string `json:"applicationVersion"`
	apiVersion         string `json:"apiVersion"`
}

func NewServer(secretStore store.Store) Server {
	return &server{
		secretStore: secretStore,
	}
}

func (s *server) Serve() error {
	router := s.Router()
	log.Debugf("Starting api server")
	return http.ListenAndServe(":"+listeningPort, router)
}

func (s *server) Router() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	s.createV1SubRouter("/v"+version.ApiVersion, router)
	s.createV1SubRouter("/latest", router)
	return router
}

func (s *server) createV1SubRouter(path string, router *mux.Router) {
	subrouter := router.PathPrefix(path + "/").Subrouter()

	// Handler for returning vernsion information
	// GET /latest/secrets/version
	// GET /v1/secrets/version
	subrouter.HandleFunc("/version", s.version).Methods("GET")

	// Handler for creating secrets
	// POST /v1/secrets/com.foo.app1.mysql
	//                Content-Type: application/json
	//                 {secret: ...}
	// POST /latest/secrets/com.foo.app1.mysql
	//                Content-Type: application/json
	//                 {secret: ...}
	subrouter.HandleFunc("/secrets/{name}", s.postSecret).Methods("POST")

	// Handler for removing secrets:
	// POST /v1/revoke/com.foo.app1.mysql/1
	// POST /latest/revoke/com.foo.app1.mysql/1
	subrouter.HandleFunc("/revoke/{name}/{serial}", s.revokeSecret).Methods("POST")

	// Handler for fetching secrets:
	// GET /v1/secrets/com.foo.app1.mysql
	// GET /latest/secrets/com.foo.app1.mysql
	subrouter.HandleFunc("/secrets/{name}", s.getSecret).Methods("GET")

	// Handler for fetching secrets with version:
	// GET /v1/secrets/com.foo.app1.mysql/2
	// GET /latest/secrets/com.foo.app1.mysql/2
	subrouter.HandleFunc("/secrets/{name}/{serial}", s.getSecret).Methods("GET")
}

func (s *server) postSecret(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars["name"]
	if request.Body == nil {
		log.Errorf("Bad data supplied for creating secret: %s", name)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(request.Body)
	var secretPayload api.SecretPayload
	err := decoder.Decode(&secretPayload)
	if err != nil {
		log.Errorf("Bad data supplied for creating secret: %s, error: %v", name, err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debugf("Creating secret: name: %s", name)
	_, err = s.secretStore.Save(&api.SecretRecord{
		Name:    name,
		Serial:  int64(1),
		Payload: secretPayload.Payload,
		Active:  true,
	})
	if err != nil {
		log.Errorf("Error creating secret for name: %s, %v", name, err)
		writer.WriteHeader(http.StatusBadRequest)
	}
}
func (s *server) revokeSecret(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars["name"]
	serial := vars["serial"]
	log.Debugf("Revoking secret: name: %s, serial: %s", name, serial)
	err := s.secretStore.Revoke(name, serial)
	if err != nil {
		log.Errorf("Error revoking secret name %s: %v", name, err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (s *server) getSecret(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars["name"]
	serial := vars["serial"]
	log.Debugf("Getting secret name: %s, serial: %s", name, serial)
	secret, err := s.secretStore.Get(name, serial)
	if err != nil {
		log.Errorf("getSecret: Error getting secret name: %s, %v", name, err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(writer)
	encoder.Encode(&secret)
}

func (s *server) version(writer http.ResponseWriter, request *http.Request) {
	log.Debugf("Returning api version: %s", version.ApiVersion)
	writer.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(writer)
	encoder.Encode(&versionResponse{
		application:        version.AppName,
		applicationVersion: version.Version,
		apiVersion:         version.ApiVersion,
	})
}
