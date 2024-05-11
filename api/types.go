package api

import (
	"bytes"
	"fmt"

	"github.com/alphabill-org/alphabill-explorer-backend/util"
)

type PubKey []byte

type PubKeyHash []byte

type TxHash []byte

func (uid TxHash) String() string {
	return fmt.Sprintf("%X", []byte(uid))
}

func (uid TxHash) Eq(id TxHash) bool {
	return bytes.Equal(uid, id)
}

func (uid TxHash) MarshalText() ([]byte, error) {
	return util.ToHex(uid), nil
}

func (uid *TxHash) UnmarshalText(src []byte) error {
	res, err := util.FromHex(src)
	if err == nil {
		*uid = res
	}
	return err
}
