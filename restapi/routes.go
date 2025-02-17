package restapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/alphabill-org/alphabill-explorer-backend/restapi/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title		Alphabill Blockchain Explorer API
// @version		1.0
// @description	API to query blocks and transactions of Alphabill
// @BasePath	/api/v1

func (api *RestAPI) Router() *mux.Router {
	// TODO add request/response headers middleware
	router := mux.NewRouter().StrictSlash(true)
	router.Use(loggerMiddleware)

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

	apiV1.HandleFunc("/search", api.search).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/round-number", api.roundNumberFunc).Methods(http.MethodGet, http.MethodOptions)

	//block
	apiV1.HandleFunc("/blocks/{blockNumber}", api.getBlock).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/partitions/{partitionID}/blocks", api.getBlocksInRange).Methods(http.MethodGet, http.MethodOptions)

	//tx
	apiV1.HandleFunc("/txs/{txHash}", api.getTx).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/partitions/{partitionID}/txs", api.getTxs).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/partitions/{partitionID}/blocks/{blockNumber}/txs", api.getBlockTxsByBlockNumber).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/units/{unitID}/txs", api.getTxsByUnitID).Methods(http.MethodGet, http.MethodOptions)

	//bill
	apiV1.HandleFunc("/address/{pubKey}/bills", api.getBillsByPubKey).Methods(http.MethodGet, http.MethodOptions)
	return router
}

func (api *RestAPI) healthRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("OK - %v", time.Now())))
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Error reading request body: %v\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reader := io.NopCloser(bytes.NewBuffer(buf))
		r.Body = reader

		fmt.Printf("Server: request from=%s to=%s:%s body=%v.\n", r.RemoteAddr, r.Method, r.RequestURI, string(buf))

		next.ServeHTTP(w, r)
	})
}
