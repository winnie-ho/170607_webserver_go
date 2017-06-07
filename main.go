package main 

import (
	"net/http"
	"context"
	"log"
	"encoding/json"

    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)


//types for product config and errors
type config struct {
	Error ErrorObj `json: "error" bson:"error"`
	ConfigBody ConfigObj `json: "configBody" bson:"configBody"`
}

type ErrorObj struct {
	Code int `json: "code" bson:"code"`
	Message string `json: "message" bson:"message"`
}

type ConfigObj struct {
	Name string `json: "name" bson:"name"`
	ProductId string `json: "productId" bson:"productId"`
}

//adding the adapter interface
type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
  for _, adapter := range adapters {
    h = adapter(h)
  }
  return h
}

//handles db session for handlers and store it in context. Returns and adapter.
func withDB(db *mgo.Session) Adapter {
  // return the Adapter
  return func(h http.Handler) http.Handler {
    // the adapter (when called) should return a new handler
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      // copy the database session      
      dbsession := db.Copy()
      defer dbsession.Close()

      // save it in the mux context with a key of "database"
      context.Set(r, "database", dbsession)
      // pass execution to the original handler
      h.ServeHTTP(w, r)
    })
  }
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
		case "GET":
	handleRead(w, r)
		case "POST":
	handleInsert(w, r)
		default:
	http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}

func main() {
	// connect to the database
	db, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal("cannot dial mongo connection", err)
	}
	defer db.Close() //close the db connection

	// Adapt our handle function using withDB
	h := Adapt(http.HandlerFunc(handle), withDB(db))

	// add the handler
	http.Handle("/config/", context.ClearHandler(h))
	http.Handle("/config/:id", context.ClearHandler(h))

	// start the server
	log.Println("Listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}


func handleInsert(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "database").(*mgo.Session)

	// decode the request body
	var c config
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// insert it into the database
	if err := db.DB("avProductConfig").C("configs").Insert(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// redirect to it
	http.Redirect(w, r, "/configs/", http.StatusTemporaryRedirect)
}

func handleRead(w http.ResponseWriter, r *http.Request) {
  db := context.Get(r, "database").(*mgo.Session)
  // load the configs
  var configs []*config
  if err := db.DB("avProductConfig").C("configs").
    Find(nil).Sort("-when").Limit(10).All(&configs); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  // write it out
  if err := json.NewEncoder(w).Encode(configs); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}

func handleReadById(w http.ResponseWriter, r *http.Request) {
  db := context.Get(r, "database").(*mgo.Session)
  // load the configs


//find how to access id from params in url line
	id := pat.Param(r, "id")



  var configs []*config
  if err := db.DB("avProductConfig").C("configs").
		Find(bson.M{"configbody.productid": id}).One(&configs); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Failed find ProductId: ", err)
    return
  }
  // write it out
  if err := json.NewEncoder(w).Encode(configs); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}



