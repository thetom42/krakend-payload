package main

import (
	"context"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

const idHeaderName = "X-Payload-Id"

type registerer string

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer("krakend-payload-router")

// ClientRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var ClientRegisterer = registerer("krakend-payload-proxy")

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, handler http.Handler) (http.Handler, error) {
	//attachUserID, ok := extra["attachuserid"].(string)
	//if !ok {
	//	panic(errors.New("incorrect config").Error())
	//}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rq := new(http.Request)
		*rq = *r

		rq.Header.Set(idHeaderName, uuid.NewString())

		fmt.Println("krakend-payload handler about to do something")

		rBodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotAcceptable)
			return
		}
		fmt.Println(string(rBodyBytes))

		handler.ServeHTTP(w, rq)
	}), nil
}

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(ctx context.Context, extra map[string]interface{}) (h http.Handler, e error) {
	// check the passed configuration and initialize the plugin
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}
	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http client
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("before proxy handler %s", html.EscapeString(req.URL.Path))
		// fmt.Println(req.URL.Query().Get("appid"))

		makeOriginalRequest(w, req)

		fmt.Println("after proxy-plugin called")
	}), nil
}

func init() {
	fmt.Println("Payload plugin loaded!")
}

func makeOriginalRequest(w http.ResponseWriter, req *http.Request) {
	client := &http.Client{}
	// Send an HTTP request and returns an HTTP response object.
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	// headers
	for name, values := range resp.Header {
		w.Header()[name] = values
	}

	// status (must come after setting headers and before copying body)
	w.WriteHeader(resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("request failed: %s\n", err)
		return
	}

	w.Write(body)

	fmt.Println(string(body))

	fmt.Println("request completed")
}

func writedata(uuid string, data []byte) {
	// You can generate a Token from the "Tokens Tab" in the UI
	const (
		token  = "FILL_IN"
		host   = "http://localhost:8086"
		org    = "tmo"
		bucket = "tmobuck"
	)

	client := influxdb2.NewClient(host, token)
	// always close client at the end
	defer client.Close()

	// get non-blocking write client
	writeAPI := client.WriteAPI(org, bucket)

	// write line protocol
	writeAPI.WriteRecord(fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0))
	writeAPI.WriteRecord(fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 22.5, 45.0))
	// Flush writes
	writeAPI.Flush()

	query := fmt.Sprintf("from(bucket:\"%v\")|> range(start: -1h) |> filter(fn: (r) => r._measurement == \"stat\")", bucket)
	// Get query client
	queryAPI := client.QueryAPI(org)
	// get QueryTableResult
	result, err := queryAPI.Query(context.Background(), query)
	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Notice when group key has changed
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// Access data
			fmt.Printf("value: %v\n", result.Record().Value())
		}
		// check for an error
		if result.Err() != nil {
			fmt.Printf("query parsing error: %s\n", result.Err().Error())
		}
	} else {
		panic(err)
	}
}

func main() {}
