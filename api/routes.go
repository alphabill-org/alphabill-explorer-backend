package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/alphabill-org/alphabill-explorer-backend/internal/log"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/alphabill-org/alphabill-explorer-backend/api/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title		Alphabill Blockchain Explorer API
// @version		1.0
// @description	API to query blocks and transactions of Alphabill
// @BasePath	/api/v1

func (c *Controller) Router() *mux.Router {
	// TODO add request/response headers middleware
	router := mux.NewRouter().StrictSlash(true)
	router.Use(loggerMiddleware)

	router.Path("/health").HandlerFunc(c.healthRequest)

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

	apiV1.HandleFunc("/search", c.search).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/round-number", c.roundNumber).Methods(http.MethodGet, http.MethodOptions)

	//block
	apiV1.HandleFunc("/blocks/{blockNumber}", c.getBlock).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/partitions/{partitionID}/blocks", c.getBlocksInRange).Methods(http.MethodGet, http.MethodOptions)

	//tx
	apiV1.HandleFunc("/txs/{txHash}", c.getTx).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/partitions/{partitionID}/txs", c.getTxs).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/partitions/{partitionID}/blocks/{blockNumber}/txs", c.getBlockTxsByBlockNumber).Methods(http.MethodGet, http.MethodOptions)
	apiV1.HandleFunc("/units/{unitID}/txs", c.getTxsByUnitID).Methods(http.MethodGet, http.MethodOptions)

	//bill
	apiV1.HandleFunc("/address/{pubKey}/bills", c.getBillsByPubKey).Methods(http.MethodGet, http.MethodOptions)
	return router
}

func (c *Controller) healthRequest(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("OK - %v", time.Now())))
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("Error reading request body", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reader := io.NopCloser(bytes.NewBuffer(buf))
		r.Body = reader

		log.Info("Request", "from", r.RemoteAddr, "to", fmt.Sprintf("%s:%s", r.Method, r.RequestURI), "body", string(buf))

		next.ServeHTTP(w, r)
	})
}
