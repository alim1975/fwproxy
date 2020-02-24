package db

import (
       "testing"
       "github.com/stretchr/testify/assert"
       "io/ioutil"
       "net/http"
       "net/http/httptest"
)

func TestAddDelete(t *testing.T) {
     dbase := New("/urlpath/1")
     dbase.Add("key1", "value1")
     assert.True(t, dbase.hasKey("key1"), "Added key, key1, not found in the DB.")
     dbase.Remove("key1")
     assert.False(t, dbase.hasKey("key1"), "Removed key, key1, is still in the DB.")
}

func TestHandleHttpGet(t *testing.T) {
     dbase := New("/urlpath/1/")
     dbase.Add("foo", "bar")
     url := "http://127.0.0.1:8000" + dbase.prefix
     // URL in DB
     req := httptest.NewRequest("GET", url + "foo", nil)
     rw := httptest.NewRecorder()
     dbase.HandleHttpGet(rw, req)
     res := rw.Result()
     assert.Equal(t, 200, res.StatusCode, "Expecting Status OK.")
     body, _ := ioutil.ReadAll(res.Body)
     res.Body.Close()
     assert.Equal(t, "bar", string(body))
     // URL not in DB      
     req = httptest.NewRequest("GET", url + "something", nil)
     rw = httptest.NewRecorder()
     dbase.HandleHttpGet(rw, req)
     res = rw.Result()
     assert.Equal(t, 200, res.StatusCode, "Expecting Status OK.")
     body, _ = ioutil.ReadAll(res.Body)
     res.Body.Close()
     assert.Equal(t, "SAFE", string(body))
     // Bad request
     req = httptest.NewRequest("GET", "http://127.0.0.1/" + "something", nil)
     rw = httptest.NewRecorder()
     dbase.HandleHttpGet(rw, req)
     res = rw.Result()
     assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Expecting Status BadRequest.")
     // Unknown method POST
     req = httptest.NewRequest("POST", url + "foo", nil)
     rw = httptest.NewRecorder()
     dbase.HandleHttpGet(rw, req)
     res = rw.Result()
     assert.Equal(t, http.StatusNotImplemented, res.StatusCode, "Expecting Status  NotImplemented.")
}
