package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"io/ioutil"
	"time"
)

const outputHeaderName = "X-Friend-User"

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer("krakend-payload")

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

//func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, _ http.Handler) (http.Handler, error) {
//	// check the passed configuration and initialize the plugin
//	name, ok := extra["name"].(string)
//	if !ok {
//		return nil, errors.New("wrong config")
//	}
//	if name != string(r) {
//		return nil, fmt.Errorf("unknown register %s", name)
//	}
//	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
//	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
//		fmt.Fprintf(w, "Hello, %q", html.EscapeString(req.URL.Path))
//	}), nil
//}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, handler http.Handler) (http.Handler, error) {
	attachUserID, ok := extra["attachuserid"].(string)
	if !ok {
		panic(errors.New("incorrect config").Error())
	}

	client := &http.Client{Timeout: 3 * time.Second}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.github.com/users/%v", attachUserID), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rq.Header.Set("Content-Type", "application/json")

		rs, err := client.Do(rq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}
		defer rs.Body.Close()

		rsBodyBytes, err := ioutil.ReadAll(rs.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}

		r2 := new(http.Request)
		*r2 = *r

		r2.Header.Set(outputHeaderName, string(rsBodyBytes))

		fmt.Println("krakend-payload handler about to do something")

		rBodyBytes, err :=  ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}
		fmt.Println(string(rBodyBytes))

		handler.ServeHTTP(w, r2)
	}), nil
}

func init() {
	fmt.Println("krakend-payload handler plugin loaded!")
}

func main() {}
