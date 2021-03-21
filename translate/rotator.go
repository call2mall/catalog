package translate

import (
	"sync"
)

type Rotator struct {
	mx    *sync.Mutex
	list  *sync.Map
	total int
	ix    int
}

func NewRotator(list []Translator) (r *Rotator) {
	r = &Rotator{
		mx:    &sync.Mutex{},
		list:  &sync.Map{},
		total: len(list),
	}

	for ix, s := range list {
		r.list.Store(ix, s)
	}

	return
}

func (r *Rotator) Next() (translator Translator, ok bool) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.total == 0 {
		return
	}

	var val interface{}
	val, ok = r.list.Load(r.ix)
	if !ok {
		val, ok = r.list.Load(0)
		if !ok {
			return
		}

		r.ix = -1
	}

	translator = val.(Translator)

	r.ix++
	if r.ix >= r.total {
		r.ix = 0
	}

	ok = true

	return
}
