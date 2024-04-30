package restapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// @Summary Retrieve bills by public key
// @Description Get bills associated with a specific public key
// @Tags bills
// @Accept json
// @Produce json
// @Param pubKey path string true "Public Key"
// @Success 200 {array} api.Bill "List of bills"
// @Failure 400 {object} ErrorResponse "Error: Missing 'pubKey' variable in the URL"
// @Failure 404 {object} ErrorResponse "Error: Bills with specified public key not found"
// @Router /address/{pubKey}/bills [get]
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