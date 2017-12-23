package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type NoneCfg struct {
}

type RePolicy struct {
	RetryOn string `json:"retry_on"`
	Nums    int    `json:"num_retries"`
	Timeout int    `json:"per_try_timeout_ms"`
}

var globalRePolicy = RePolicy{RetryOn: "connect-failure", Nums: 3, Timeout: 100}

type Route struct {
	Timeout uint32   `json:"timeout_ms"`
	Prefix  string   `json:"prefix"`
	Cluster string   `json:"cluster"`
	Retry   RePolicy `json:"retry_policy"`
}

type Virtualhost struct {
	Name    string   `json:"name"`
	Domains []string `json:"domains"`
	Routes  []Route  `json:"routes"`
}

type RouteConfig struct {
	Virtual_hosts []Virtualhost `json:"virtual_hosts"`
}

type RDS struct {
	Cluster           string `json:"cluster"`
	Route_config_name string `json:"route_config_name"`
	Refresh_delay_ms  uint32 `json:"refresh_delay_ms"`
}

type Filter struct {
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type httpConfig struct {
	Codec_type  string   `json:"codec_type"`
	Stat_prefix string   `json:"stat_prefix"`
	Rds         RDS      `json:"rds"`
	Filters     []Filter `json:"filters"`
}

type Listener struct {
	Address string   `json:"address"`
	Filters []Filter `json:"filters"`
}

type Listeners struct {
	List []Listener `json:"listeners"`
}

var globalRDS = RDS{Cluster: "xds_service", Route_config_name: "default", Refresh_delay_ms: 10000}

// GET /v1/listeners/(string: service_cluster)/(string: service_node)
func getListenersCfg(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	log.Println("Path:", r.URL.Path)
	log.Println("Request:", r)

	route := Filter{Name: "router", Config: NoneCfg{}}

	httpcfg := &httpConfig{Codec_type: "auto", Stat_prefix: "ingress_http", Rds: globalRDS}
	httpcfg.Filters = make([]Filter, 0)
	httpcfg.Filters = append(httpcfg.Filters, route)

	httpfilter := Filter{Name: "http_connection_manager", Config: httpcfg}

	address := "tcp://0.0.0.0:8080"

	listener := Listener{Address: address, Filters: make([]Filter, 0)}
	listener.Filters = append(listener.Filters, httpfilter)

	listeners := Listeners{List: make([]Listener, 0)}
	listeners.List = append(listeners.List, listener)

	body, err := json.Marshal(listeners)
	if err != nil {
		log.Println(err.Error())

	} else {
		log.Printf("%s", string(body))
		fmt.Fprintf(w, "%s", string(body))
	}
}

// GET /v1/routes/(string: route_config_name)/(string: service_cluster)/(string: service_node)
func getRoutesCfg(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	log.Println("Path:", r.URL.Path)
	log.Println("Request:", r)

	route1 := Route{Timeout: 0, Prefix: "/service/1", Cluster: "service1", Retry: globalRePolicy}
	route2 := Route{Timeout: 0, Prefix: "/service/2", Cluster: "service2", Retry: globalRePolicy}

	virtualhost := Virtualhost{Name: "backend", Domains: []string{"*"}, Routes: make([]Route, 0)}
	virtualhost.Routes = append(virtualhost.Routes, route1)
	virtualhost.Routes = append(virtualhost.Routes, route2)

	routeConfig := RouteConfig{Virtual_hosts: []Virtualhost{virtualhost}}

	body, err := json.Marshal(routeConfig)
	if err != nil {
		log.Println(err.Error())

	} else {
		log.Printf("%s", string(body))
		fmt.Fprintf(w, "%s", string(body))
	}
}

type HealthCheck struct {
	Type       string `json:"type"`
	Timeout    int    `json:"timeout_ms"`
	IntervalMs int    `json:"interval_ms"`
	Unhealthy  int    `json:"unhealthy_threshold"`
	Headlthy   int    `json:"healthy_threshold"`
	Path       string `json:"path"`
}

var globalHC = HealthCheck{Type: "http", Timeout: 250, IntervalMs: 1000, Unhealthy: 5, Headlthy: 5, Path: "/healthcheck"}

type Cluster struct {
	Name    string      `json:"name"`
	Timeout uint32      `json:"connect_timeout_ms"`
	Type    string      `json:"type"`
	LbType  string      `json:"lb_type"`
	Hc      HealthCheck `json:"health_check"`
}

type Clusters struct {
	ClusterList []Cluster `json:"clusters"`
}

// GET /v1/clusters/(string: service_cluster)/(string: service_node)
func getClustersCfg(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	log.Println("Path:", r.URL.Path)
	log.Println("Request:", r)

	service1 := Cluster{Name: "service1", Timeout: 250, Type: "sds", LbType: "round_robin", Hc: globalHC}
	service2 := Cluster{Name: "service2", Timeout: 250, Type: "sds", LbType: "random", Hc: globalHC}

	clusters := Clusters{[]Cluster{service1, service2}}

	body, err := json.Marshal(clusters)
	if err != nil {
		log.Println(err.Error())

	} else {
		log.Printf("%s", string(body))
		fmt.Fprintf(w, "%s", string(body))
	}
}

type EndPoint struct {
	IP   string `json:"ip_address"`
	Port uint32 `json:"port"`
}

type Hosts struct {
	EdpList []EndPoint `json:"hosts"`
}

func (h *Hosts) Add(edp EndPoint) bool {
	for _, vedp := range h.EdpList {
		if vedp.IP == edp.IP &&
			vedp.Port == edp.Port {
			return false
		}
	}

	h.EdpList = append(h.EdpList, edp)
	return true
}

var globalEndPoint = make(map[string]*Hosts, 100)

// GET /v1/registration/(string: service_name)
func getRegisterCfg(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	log.Println("Path:", r.URL.Path)
	log.Println("Request:", r)

	service_name := strings.TrimPrefix(r.URL.Path, "/v1/registration/")

	hosts, bflag := globalEndPoint[service_name]
	if bflag == false {
		log.Printf("%s end point can not found!", service_name)
		http.NotFound(w, r)
		return
	}

	body, err := json.Marshal(hosts)
	if err != nil {
		log.Println(err.Error())

	} else {
		log.Printf("%s", string(body))
		fmt.Fprintf(w, "%s", string(body))
	}
}

func readFully(conn io.ReadCloser) ([]byte, error) {
	result := bytes.NewBuffer(nil)
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		result.Write(buf[0:n])
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
	}
	return result.Bytes(), nil
}

// POST /v2/register/(string: service_name)
func postRegisterCfg(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	//log.Println("Path:", r.URL.Path)

	service_name := strings.TrimPrefix(r.URL.Path, "/v2/register/")

	if len(service_name) == 0 {
		log.Println("service name is invaled!")
		http.NotFound(w, r)
		return
	}

	hosts, bflag := globalEndPoint[service_name]
	if bflag == false {
		hosts = &Hosts{EdpList: make([]EndPoint, 0)}
		globalEndPoint[service_name] = hosts
	}

	body, err := readFully(r.Body)
	if err != nil {
		log.Println("read body failed!", err.Error())
		http.NotFound(w, r)
		return
	}

	endpoint := new(EndPoint)

	err = json.Unmarshal(body, endpoint)
	if err != nil {
		log.Println("json Unmarshal body failed!", err.Error(), string(body))
		http.NotFound(w, r)
		return
	}

	if hosts.Add(*endpoint) {
		log.Printf("%s add endpoint [%s:%d] success!", service_name, endpoint.IP, endpoint.Port)
	}

	fmt.Fprintf(w, "%s", "add endpoint success!")
}

func main() {

	http.HandleFunc("/v2/register/", postRegisterCfg)

	http.HandleFunc("/v1/listeners/", getListenersCfg)
	http.HandleFunc("/v1/routes/", getRoutesCfg)
	http.HandleFunc("/v1/clusters/", getClustersCfg)
	http.HandleFunc("/v1/registration/", getRegisterCfg)

	err := http.ListenAndServe(":1001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
