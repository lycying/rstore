package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/lycying/rstore/cfg"
	"io/ioutil"
	"net/http"
	"runtime/debug"
	"time"
)

var ResponseOK = []byte("{\"result\":\"OK\"}")

func ResponseError(err error) []byte {
	resp := make(map[string]string)
	resp["result"] = err.Error()
	resp["msg"] = string(debug.Stack())
	b, _ := json.Marshal(resp)
	return b
}

func Start() {
	api := NewApi()

	r := mux.NewRouter()
	r.PathPrefix("/plugins/").Handler(http.StripPrefix("/plugins/", http.FileServer(http.Dir("plugins"))))

	r.HandleFunc("/api/db/save", api.db_save).Methods("POST", "PUT")
	r.HandleFunc("/api/db/all/{type}", api.db_all).Methods("GET")
	r.HandleFunc("/api/db/delete/{type}/{name}", api.db_delete).Methods("DELETE")

	r.HandleFunc("/api/dbgroup/save", api.dbgroup_save).Methods("POST", "PUT")
	r.HandleFunc("/api/dbgroup/all", api.dbgroup_all).Methods("GET")
	r.HandleFunc("/api/dbgroup/delete/{name}", api.dbgroup_delete).Methods("DELETE")

	r.HandleFunc("/api/shard/save", api.shard_save).Methods("POST", "PUT")
	r.HandleFunc("/api/shard/all", api.shard_all).Methods("GET")
	r.HandleFunc("/api/shard/delete/{name}", api.shard_delete).Methods("DELETE")

	r.HandleFunc("/api/rule/save", api.rule_save).Methods("POST", "PUT")
	r.HandleFunc("/api/rule/all", api.rule_all).Methods("GET")
	r.HandleFunc("/api/rule/delete/{name}", api.rule_delete).Methods("DELETE")

	http.Handle("/", r)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8888",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	srv.ListenAndServe()
}

type Api struct {
	saver cfg.Saver
	fly cfg.Saver
}

func NewApi() *Api {
	api := &Api{}
	api.saver = cfg.GetSaver()
	api.fly = cfg.GetFly()
	return api
}

func (api *Api) dbgroup_save(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := ioutil.ReadAll(r.Body)

	var item *cfg.CfgDBGroup
	err := json.Unmarshal(b, &item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	err = api.fly.SaveOrUpdateDBGroup(item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	err = api.saver.SaveOrUpdateDBGroup(item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(ResponseOK)
}


func (api *Api) db_save(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := ioutil.ReadAll(r.Body)

	var objmap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objmap)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}

	typeB, _ := objmap["Type"].MarshalJSON()
	if string(typeB) == "\"postgres\"" {
		var item *cfg.CfgDBPostgres
		err = json.Unmarshal(b, &item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.fly.SaveOrUpdatePostgres(item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.saver.SaveOrUpdatePostgres(item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
	} else if string(typeB) == "\"mysql\"" {
		var item *cfg.CfgDBMysql
		err = json.Unmarshal(b, &item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.fly.SaveOrUpdateMySql(item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.saver.SaveOrUpdateMySql(item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
	} else if string(typeB) == "\"redis\"" {
		var item *cfg.CfgDBRedis
		err = json.Unmarshal(b, &item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.fly.SaveOrUpdateRedis(item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.saver.SaveOrUpdateRedis(item)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
	}
	w.Write(ResponseOK)
}

func (api *Api) db_delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	t := vars["type"]
	n := vars["name"]

	if t == "pg" {
		err := api.fly.RemovePostgres(n)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.saver.RemovePostgres(n)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
	} else if t == "mysql" {
		err := api.fly.RemoveMysql(n)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.saver.RemoveMysql(n)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}

	} else if t == "redis" {
		err := api.fly.RemoveRedis(n)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
		err = api.saver.RemoveRedis(n)
		if err != nil {
			w.Write(ResponseError(err))
			return
		}
	}
	w.Write(ResponseOK)
}

func (api *Api) shard_save(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := ioutil.ReadAll(r.Body)

	var item *cfg.CfgShard
	err := json.Unmarshal(b, &item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	err = api.fly.SaveOrUpdateShard(item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	err = api.saver.SaveOrUpdateShard(item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(ResponseOK)
}


func (api *Api) dbgroup_delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	n := vars["name"]

	err := api.fly.RemoveDBGroup(n)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	err = api.saver.RemoveDBGroup(n)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(ResponseOK)
}
func (api *Api) shard_delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	n := vars["name"]

	err := api.fly.RemoveShard(n)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	err = api.saver.RemoveShard(n)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(ResponseOK)
}

func (api *Api) rule_save(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := ioutil.ReadAll(r.Body)

	var item *cfg.CfgRule
	err := json.Unmarshal(b, &item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	err = api.fly.SaveOrUpdateRule(item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	err = api.saver.SaveOrUpdateRule(item)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(ResponseOK)
}

func (api *Api) rule_delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	n := vars["name"]

	err := api.fly.RemoveRule(n)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}

	err = api.saver.RemoveRule(n)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(ResponseOK)
}
func (api *Api) dbgroup_all(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	all, err := api.saver.GetAllDBGroup()
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	b, err := json.Marshal(all)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(b)
}

func (api *Api) db_all(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	t := vars["type"]
	var err error
	var all interface{}
	if t == "pg" {
		all, err = api.saver.GetAllPostgres()
	} else if t == "mysql" {
		all, err = api.saver.GetAllMysql()
	} else if t == "redis" {
		all, err = api.saver.GetAllRedis()
	}
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	b, err := json.Marshal(all)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(b)
}

func (api *Api) shard_all(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	all, err := api.saver.GetAllShard()
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	b, err := json.Marshal(all)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(b)
}
func (api *Api) rule_all(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	all, err := api.saver.GetAllRule()
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	b, err := json.Marshal(all)
	if err != nil {
		w.Write(ResponseError(err))
		return
	}
	w.Write(b)
}
