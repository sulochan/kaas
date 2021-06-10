package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"github.com/sulochan/kaas/api"
)

func main() {
	chain := alice.New()
	router := mux.NewRouter()
	http.Handle("/", router)
	//http.Handle("/static/", chain.ThenFunc(web.HandleStatic))

	// UI routes
	//routes.HandleUIRoutes(router)

	// API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Handle("/clusters", chain.Append(api.SetContext).ThenFunc(api.GetAllClusters)).Methods("GET")
	apiRouter.Handle("/clusters/{cluster:[[A-Z,a-z,0-9,-]+}", chain.Append(api.SetContext).ThenFunc(api.GetCluster)).Methods("GET")
	apiRouter.Handle("/clusters/{cluster:[A-Z,a-z,0-9,-]+}/nodes", chain.Append(api.SetContext).ThenFunc(api.GetClusterNodes)).Methods("GET")

	apiRouter.Handle("/clusters", chain.Append(api.SetContext).ThenFunc(api.CreateCluster)).Methods("POST")
	apiRouter.Handle("/clusters/{cluster:[[A-Z,a-z,0-9,-]+}", chain.Append(api.SetContext).ThenFunc(api.UpdateCluster)).Methods("POST")

	apiRouter.Handle("/clusters/{cluster:[A-Z,a-z,0-9,-]+}", chain.Append(api.SetContext).ThenFunc(api.DeleteCluster)).Methods("DELETE")
	apiRouter.Handle("/clusters/{cluster:[[A-Z,a-z,0-9,-]+}/nodes/{node:[A-Z,a-z,0-9,-]+}", chain.Append(api.SetContext).ThenFunc(api.DeleteClusterNode)).Methods("DELETE")

	if err := http.ListenAndServe(fmt.Sprintf(":%s", "9191"), nil); err != nil {
		log.Info("http.ListendAndServer() failed with %s\n", err)
	}
	log.Info("Exited\n")
}
