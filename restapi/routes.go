package restapi

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func (api *MoneyRestAPI) Router() *mux.Router {
	// TODO add request/response headers middleware
	router := mux.NewRouter().StrictSlash(true)

	apiRouter := router.PathPrefix("/api").Subrouter()
	// add cors middleware
	// content-type needs to be explicitly defined without this content-type header is not allowed and cors filter is not applied
	// Link header is needed for pagination support.
	// OPTIONS method needs to be explicitly defined for each handler func
	apiRouter.Use(handlers.CORS(
		handlers.AllowedHeaders([]string{ContentType}),
		handlers.ExposedHeaders([]string{HeaderLink}),
	))

	// version v1 router
	apiV1 := apiRouter.PathPrefix("/v1").Subrouter()

	apiV1.HandleFunc("/blocks/{blockNumber}", api.getBlock).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/blocks", api.getBlocks).Methods("GET", "OPTIONS")

	apiV1.HandleFunc("/txs/{txHash}", api.getTx).Methods("GET", "OPTIONS")
	apiV1.HandleFunc("/blocks/{blockNumber}/txs", api.getBlockTxsByBlockNumber).Methods("GET", "OPTIONS")

	return router
}
