package gob

import (
	"encoding/gob"
	"go-network/chapter12/housework"
	"io"
)

func Load(r io.Reader) ([]*housework.Chore, error) {
	var chores []*housework.Chore

	return chores, gob.NewDecoder(r).Decode(&chores)

}

func Flush(w io.Writer, chores []*housework.Chore) error {
	return gob.NewEncoder(w).Encode(chores)
}
