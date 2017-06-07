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
	ID bson.ObjectId `json: "id" bson:"_id"`
	ErrorObj string `json: "errorObj" bson:"errorObj"`
	ConfigBody string `json: "configBody" bson:"configBody"`
}

// type Error struct {
// 	code int `json: "code"`
// 	message string `json: "message"`
// }

// type Config struct {
// 	name string `json: "name"`
// 	body string `json: "body"`
// }

type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
  for _, adapter := range adapters {
    h = adapter(h)
  }
  return h
}

func withDB(db *mgo.Session) Adapter {
  // return the Adapter
  return func(h http.Handler) http.Handler {
    // the adapter (when called) should return a new handler
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      // copy the database session      
      dbsession := db.Copy()
      defer dbsession.Close() // clean up 
      // save it in the mux context
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
  http.Handle("/config", context.ClearHandler(h))
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
  // give the config a unique ID and set the time
  c.ID = bson.NewObjectId()
  // insert it into the database
  if err := db.DB("avProductConfig").C("configs").Insert(&c); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }
  // redirect to it
  http.Redirect(w, r, "/configs/"+c.ID.Hex(), http.StatusTemporaryRedirect)
}

func handleRead(w http.ResponseWriter, r *http.Request) {
  db := context.Get(r, "database").(*mgo.Session)
  // load the configs
  var configs []*config
  if err := db.DB("avProductConfig").C("configs").
    Find(nil).Sort("-when").Limit(100).All(&configs); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  // write it out
  if err := json.NewEncoder(w).Encode(configs); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}


// func config(w http.ResponseWriter, r *http.Request) {
// 	io.WriteString(w, "This is config root page")
// }

// func main() {
// 	//getting a session
// 	session, err := mgo.Dial("localhost")
// 	if err != nil {
// 		panic(err)
// 	}
// 	//closing session after use
// 	defer session.Close()

// 	session.SetMode(mgo.Monotonic, true)
// 	ensureIndex(session)


// 	mux := goji.NewMux()
// 	mux.HandleFunc(pat.Get("/config"), allConfigs(session))
// 	mux.HandleFunc(pat.Post("/configs"), addConfig(session))
// 	mux.HandleFunc(pat.Get("/configs/:id"), configById(session))
// 	mux.HandleFunc(pat.Put("/configs/:id"), updateConfig(session))

// 	log.Println("Listening on port 8080")
// 	http.ListenAndServe(":8080", mux)
// }

// func ensureIndex(s *mgo.Session) {
// 	session := s.Copy()
// 	defer session.Close()

// 	c := session.DB("store").C("books")

// 	index := mgo.Index{
// 		Key: []string{"id"},
// 		Unique: true,
// 		DropDups: true,
// 		Background: true,
// 		Sparse: true,
// 	}
// 	err := c.EnsureIndex(index)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func allConfigs(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {  
//     return func(w http.ResponseWriter, r *http.Request) {
//         session := s.Copy()
//         defer session.Close()

//         c := session.DB("store").C("configs")

//         var configs []Config
//         err := c.Find(bson.M{}).All(&configs)
//         if err != nil {
//             ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
//             log.Println("Failed to get all configs: ", err)
//             return
//         }

//         respBody, err := json.MarshalIndent(configs, "", "  ")
//         if err != nil {
//             log.Fatal(err)
//         }

//         ResponseWithJSON(w, respBody, http.StatusOK)
//     }
// }

// func addConfig(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {  
//     return func(w http.ResponseWriter, r *http.Request) {
//         session := s.Copy()
//         defer session.Close()

//         var config Config
//         decoder := json.NewDecoder(r.Body)
//         err := decoder.Decode(&config)
//         if err != nil {
//             ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
//             return
//         }

//         c := session.DB("store").C("configs")

//         err = c.Insert(config)
//         if err != nil {
//             if mgo.IsDup(err) {
//                 ErrorWithJSON(w, "Config with this ID already exists", http.StatusBadRequest)
//                 return
//             }

//             ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
//             log.Println("Failed to insert product config: ", err)
//             return
//         }

//         w.Header().Set("Content-Type", "application/json")
//         w.Header().Set("Location", r.URL.Path+"/"+config.Id)
//         w.WriteHeader(http.StatusCreated)
//     }
// }

