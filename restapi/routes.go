package restapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/alphabill-org/alphabill-explorer-backend/restapi/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title		Alphabill Blockchain Explorer API
// @version		1.0
// @description	API to query blocks and transactions of Alphabill's Money Partition
// @BasePath	/api/v1

func (api *MoneyRestAPI) Router() *mux.Router {
	// TODO add request/response headers middleware
	router := mux.NewRouter().StrictSlash(true)

	router.Path("/health").HandlerFunc(api.healthRequest)

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

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

func (api *MoneyRestAPI) healthRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("OK - %v", time.Now())))
}
