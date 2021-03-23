package user_agent

import (
	"sync"
)

type Rotator struct {
	mx    *sync.Mutex
	list  *sync.Map
	total int
	ix    int
}

var (
	once     = &sync.Once{}
	mx       = &sync.Mutex{}
	instance *Rotator
)

func GetInstance() (r *Rotator) {
	mx.Lock()
	defer mx.Unlock()

	once.Do(func() {
		instance = NewRotator(GetAllUserAgentList())
	})

	r = instance

	return
}

func NewRotator(list []UserAgent) (r *Rotator) {
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

func (r *Rotator) Next() (userAgent UserAgent, ok bool) {
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

	userAgent = val.(UserAgent)

	r.ix++
	if r.ix >= r.total {
		r.ix = 0
	}

	ok = true

	return
}
