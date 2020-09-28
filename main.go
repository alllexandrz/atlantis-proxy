package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)
type DebugTransport struct{}

type BB_payload struct {
	PullRequest struct{
		ToRef struct{
			DisplayId string
		}
	}
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/",handleRequestAndRedirect )
	http.Handle("/",router)

	fmt.Println("Server is listening...")
	http.ListenAndServe(":5000", nil)
}

func (DebugTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequestOut(r, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return http.DefaultTransport.RoundTrip(r)
}

func serveReverseProxy(target string,path string,res http.ResponseWriter, req *http.Request){
	url, _ := url.Parse(target)
	proxy:= httputil.NewSingleHostReverseProxy(url)
	req.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host
	req.URL.Path = path		//"/atlantis/events"

	proxy.Transport = DebugTransport{}
	proxy.ServeHTTP(res,req)

}

func handleRequestAndRedirect(res http.ResponseWriter, req *http.Request) {
	requestPayload := parseRequestBody(req)
	url,path := getProxyUrl(requestPayload.PullRequest.ToRef.DisplayId)

	logRequestPayload(requestPayload, url)

	serveReverseProxy(url,path, res, req)
}

func getProxyUrl(proxyConditionRaw string) (host_url string,path string) {

	proxyCondition := strings.ToUpper(proxyConditionRaw)
	proxyCondition = strings.Replace(proxyCondition,"-","_",-1)

	condtion_url := os.Getenv("C_"+proxyCondition)

	if condtion_url != "" {
		u,_ := url.Parse(condtion_url)
		fmt.Printf("Founded condition %s with path %s ", u.Host, u.Path)
		return u.Scheme+"://"+u.Host,u.Path

	}else {
		return "default_host_url","/no-route"
	}

}


func logRequestPayload(requestionPayload BB_payload, proxyUrl string) {
	log.Printf("branche: %s, proxy_url: %s\n", requestionPayload.PullRequest.ToRef.DisplayId, proxyUrl)
}

func requestBodyDecoder(request *http.Request) *json.Decoder {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
		panic(err)
	}

	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return json.NewDecoder(ioutil.NopCloser(bytes.NewBuffer(body)))
}

func parseRequestBody(request *http.Request) BB_payload {
	decoder := requestBodyDecoder(request)

	var requestPayload BB_payload
	err := decoder.Decode(&requestPayload)

	if err != nil {
		panic(err)
	}

	return requestPayload
}
