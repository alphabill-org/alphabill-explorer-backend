package api

import (
	"fmt"
	"net/http"

	"github.com/alphabill-org/alphabill-explorer-backend/domain"
	"github.com/alphabill-org/alphabill-explorer-backend/util"
	"github.com/gorilla/mux"
)

// @Summary Retrieve bills by public key
// @Description Get bills associated with a specific public key
// @Tags Bills
// @Accept json
// @Produce json
// @Param pubKey path string true "Public Key"
// @Success 200 {array} domain.Bill "List of bills"
// @Failure 400 {object} ErrorResponse "Error: Missing 'pubKey' variable in the URL"
// @Failure 404 {object} ErrorResponse "Error: Bills with specified public key not found"
// @Router /address/{pubKey}/bills [get]
func (c *Controller) getBillsByPubKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pubKeyStr, ok := vars[paramPubKey]
	if !ok {
		c.rw.WriteMissingParamResponse(w, paramPubKey)
		return
	}

	pubKeyHash, err := util.PubKeyHash(pubKeyStr)
	if err != nil {
		c.rw.WriteInvalidParamResponse(w, paramPubKey)
		return
	}

	bills, err := c.MoneyService.GetBillsByPubKeyHash(r.Context(), pubKeyHash)
	if err != nil {
		c.rw.WriteInternalErrorResponse(w, fmt.Errorf("failed to load bills with pubKey %s : %w", pubKeyStr, err))
		return
	}

	if len(bills) == 0 {
		c.rw.WriteErrorResponse(w, fmt.Errorf("bills with pubKey %s not found", pubKeyStr), http.StatusNotFound)
		return
	}

	var response = []domain.Bill{}
	for _, bill := range bills {
		response = append(response, domain.Bill{
			NetworkID:   bill.NetworkID,
			PartitionID: bill.PartitionID,
			ID:          bill.ID,
			Value:       bill.Value,
			LockStatus:  bill.LockStatus,
			Counter:     bill.Counter,
		})
	}

	c.rw.WriteResponse(w, response)
}
