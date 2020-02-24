package db

import (
       "fmt"
       "github.com/sirupsen/logrus"
       "net/http"
       "strings"
)

type Db struct {
     prefix string        // URI prefix
     db map[string]string // request URLs to filter
}

var (
    logger *logrus.Logger
)

func init() {
     logrus.SetLevel(logrus.DebugLevel)
     logger = logrus.StandardLogger()
}

func New(prefix string) *Db {
     db := &Db{prefix: prefix, db: make(map[string]string)}
     return db
}

func (db *Db) Add(key, value string) {
     db.db[key] = value
}

func (db *Db) Remove(key string) {
     delete(db.db, key)
}

func (db *Db) hasKey(key string) bool {
     _, ok := db.db[key]
     return ok
}

func (db *Db) HandleHttpGet(res http.ResponseWriter, req *http.Request) {
     logger.Debug("Got request method: ", req.Method, " URL: ", req.URL)
     switch req.Method {
     case "GET":
     	  if strings.HasPrefix(req.URL.RequestURI(), db.prefix) {
     	     toks := strings.SplitAfter(req.URL.RequestURI(), db.prefix)
     	     if value, ok := db.db[toks[1]]; ok {
	     	logger.Debug("URL ", toks[1], " found in the database. value: ", value)
	     	fmt.Fprintf(res, value)
	     } else {
	       logger.Debug("URL ", toks[1], " not found in the database.")
	       fmt.Fprintf(res, "SAFE")
	     }
	  } else {
	    err := fmt.Sprintf("Invalid request: %s, expecting prefix: %s",
	    		req.URL.RequestURI(), db.prefix)
	    logger.Warn(err)
	    http.Error(res, err, http.StatusBadRequest)
	  }
     default:
	  err := fmt.Sprintf("We do not handle HTTP method: %s", req.Method)
	  logger.Warn(err)
	  http.Error(res, err, http.StatusNotImplemented)
     }
}
