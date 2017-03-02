package postgres

import "testing"

func TestPostgres_Init(t *testing.T) {
	pg, err := NewPostgres("postgres://postgres:postgres@localhost/postgres?sslmode=disable")
	if err != nil {

	}
	pg.SET("kv", "__rstore__")
	v, err := pg.GET("kv")
	if err != nil {
		t.Error(err)
	}
}
