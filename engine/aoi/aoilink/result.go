package aoilink

import (
	"sync"

	"github.com/xhaoh94/gox/engine/types"
)

var (
	pool map[any]bool
	mux  sync.Mutex
)

type AOIResult[T types.AOIKey] struct {
	idMap map[T]bool
}

func newResult[T types.AOIKey]() *AOIResult[T] {
	mux.Lock()
	defer mux.Unlock()
	if len(pool) > 0 {
		for k, v := range pool {
			delete(pool, v)
			return (k.(*AOIResult[T]))
		}
	}
	r := &AOIResult[T]{}
	r.idMap = make(map[T]bool)
	return r
}
func (r *AOIResult[T]) Has(id T) bool {
	return r.idMap[id]
}
func (r *AOIResult[T]) push(id T) {
	r.idMap[id] = true
}

func (r *AOIResult[T]) IDList() []T {
	ids := make([]T, 0)
	for id := range r.idMap {
		ids = append(ids, id)
	}
	return ids
}
func (r *AOIResult[T]) IDMap() map[T]bool {
	return r.idMap
}
func (r *AOIResult[T]) Range(call func(T)) {
	for id := range r.idMap {
		call(id)
	}
}

func (r *AOIResult[T]) Compare(cResult types.IAOIResult[T]) (Complement []T, Minus []T, Intersect []T) {
	cMap := cResult.IDMap()
	for id := range cMap {
		if r.idMap[id] {
			Intersect = append(Intersect, id)
		} else {
			Minus = append(Minus, id)
		}
	}
	for id := range r.idMap {
		if !cMap[id] {
			Complement = append(Complement, id)
		}
	}
	return
}

func (r *AOIResult[T]) Reset() {
	clear(r.idMap)
	mux.Lock()
	defer mux.Unlock()
	pool[r] = true
}
