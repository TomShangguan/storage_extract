package trie

import "github.com/ethereum/go-ethereum/rlp"

func (n *fullNode) encode(w rlp.EncoderBuffer) {
	offset := w.List()
	for _, c := range n.Children {
		if c != nil {
			c.encode(w)
		} else {
			w.Write(rlp.EmptyString)
		}
	}
	w.ListEnd(offset)
}

func (n *shortNode) encode(w rlp.EncoderBuffer) {
	offset := w.List()
	w.WriteBytes(n.Key)
	if n.Val != nil {
		n.Val.encode(w)
	} else {
		w.Write(rlp.EmptyString)
	}
	w.ListEnd(offset)
}

func (n hashNode) encode(w rlp.EncoderBuffer) {
	w.WriteBytes(n)
}

func (n valueNode) encode(w rlp.EncoderBuffer) {
	w.WriteBytes(n)
}
