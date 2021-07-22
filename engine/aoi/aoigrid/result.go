package aoigrid

import (
	"sync"

	"github.com/xhaoh94/gox/engine/aoi/aoibase"
)

var resultPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return &AOIResult{}
	},
}

type AOIResult struct {
	idMap map[string]bool
}

func newResult() *AOIResult {
	r := resultPool.Get().(*AOIResult)
	r.idMap = make(map[string]bool)
	return r
}
func (r *AOIResult) get(id string) bool {
	return r.idMap[id]
}
func (r *AOIResult) push(id string) {
	r.idMap[id] = true
}

func (r *AOIResult) IDList() []string {
	ids := make([]string, 0)
	for id := range r.idMap {
		ids = append(ids, id)
	}
	return ids
}
func (r *AOIResult) IDMap() map[string]bool {
	return r.idMap
}
func (r *AOIResult) Range(call func(string)) {
	for id := range r.idMap {
		call(id)
	}
}

func (r *AOIResult) Compare(cResult aoibase.IAOIResult) (Complement []string, Minus []string, Intersect []string) {
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

func (r *AOIResult) Reset() {
	r.idMap = nil
	resultPool.Put(r)
}
