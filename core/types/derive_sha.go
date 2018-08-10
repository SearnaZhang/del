package types
import (
	"bytes"
	"github.com/DEL-ORG/del/common"
	"github.com/DEL-ORG/del/rlp"
	"github.com/DEL-ORG/del/trie"
)
type DerivableList interface {
	Len() int
	GetRlp(i int) []byte
}
func DeriveSha(list DerivableList) common.Hash {
	keybuf := new(bytes.Buffer)
	trie := new(trie.Trie)
	for i := 0; i < list.Len(); i++ {
		keybuf.Reset()
		rlp.Encode(keybuf, uint(i))
		trie.Update(keybuf.Bytes(), list.GetRlp(i))
	}
	return trie.Hash()
}
