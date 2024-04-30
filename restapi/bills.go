package restapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (api *MoneyRestAPI) getBillsByPubKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ownerIDStr, ok := vars["pubKey"]
	if !ok {
		http.Error(w, "Missing 'pubKey' variable in the URL", http.StatusBadRequest)
		return
	}

	ownerID:= []byte(ownerIDStr)

	bills, err := api.Service.GetBillsByPubKey(r.Context() , ownerID)
	if err != nil {
		api.rw.WriteErrorResponse(w, fmt.Errorf("failed to load bills with pubKey %s : %w", ownerIDStr, err))
		return
	}

	if bills == nil {
		api.rw.ErrorResponse(w, http.StatusNotFound, fmt.Errorf("bills with pubKey %s not found", ownerIDStr))
		return
	}

	api.rw.WriteResponse(w, bills)
}