package route

import (
	"github.com/lycying/rstore/redisx"
	"github.com/lycying/rstore/redisx/postgres"
)

type EndPointer interface {
	Real() bool
}

type RealUnit struct {
	EndPointer
	Name    string
	Type    string
	Desc    string
	URL     string
	Backend redisx.Redis
}

func (unit *RealUnit) Real() bool {
	return true
}

type AbstractUnit struct {
	EndPointer
	Rules []Rule
}

func (unit *AbstractUnit) Real() bool {
	return false
}

type Router struct {
}

func NewRouter() *Router {
	router := &Router{}
	return router
}

func (router *Router) GetDBUnit(key string) (*RealUnit, error) {
	pg, _ := postgres.NewPostgres("postgres://postgres:postgres@localhost/postgres?sslmode=disable")
	r := NewSepRule(key, ":", 0, []EndPointer{
		&RealUnit{
			Backend: pg,
		},
	})
	db, err := r.RoundRobin()
	if err != nil {
	}
	if db.Real() {
		return db.(*RealUnit), nil
	}
	return nil, nil
}
