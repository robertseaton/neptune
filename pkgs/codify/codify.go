package codify

import (
	"crypto/sha256"
	"encoding/hex"
)

// TODO add salting
func SHA(str string) string {

	var bytes []byte
	bytes = []byte(str)

	h := sha256.New()                   // new sha256 object
	h.Write(bytes)                      // data is now converted to hex
	code := h.Sum(nil)                  // code is now the hex sum
	codestr := hex.EncodeToString(code) // converts hex to string

	return codestr

}
