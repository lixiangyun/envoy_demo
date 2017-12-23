package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type EndPoint struct {
	IP   string `json:"ip_address"`
	Port int    `json:"port"`
}

func (e *EndPoint) Get() string {
	return fmt.Sprintf("%s:%d", e.IP, e.Port)
}

func (e *EndPoint) Parse(addr string) (err error) {
	args := strings.Split(addr, ":")

	e.IP = args[0]
	e.Port, err = strconv.Atoi(args[1])

	return
}

var globalEndPoint EndPoint

var globalServiceName string

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	log.Println("Path:", r.URL.Path)
	log.Println("Request:", r)

	fmt.Fprintf(w, "Hello World! Form %s host[%s]! \r\n", globalServiceName, globalEndPoint.Get())
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	log.Println("Path:", r.URL.Path)
	log.Println("Request:", r)

	io.Copy(w, r.Body)
}

func ReadFully(conn io.ReadCloser) ([]byte, error) {
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

func HttpRequest(method string, url string, req []byte) (rsp []byte, err error) {
	body := bytes.NewBuffer(req)

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	rspon, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer rspon.Body.Close()

	rsp, err = ReadFully(rspon.Body)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func RegisterEndPoint() {
	for {
		body, err := json.Marshal(globalEndPoint)
		if err != nil {
			log.Println(err.Error())
			os.Exit(1)
			return
		}

		urlpath := fmt.Sprintf("http://128.5.0.20:1001/v2/register/%s", globalServiceName)

		//fmt.Println(urlpath)

		_, err = HttpRequest("POST", urlpath, body)
		if err != nil {
			log.Println(err.Error())
		} else {
			//log.Println("RSP: ", string(rsp))
		}

		time.Sleep(5 * time.Second)
	}
}

func main() {
	args := os.Args

	if len(args) < 3 {
		fmt.Println("Usage : <IP:PORT> <servicename>")
		return
	}

	globalEndPoint.Parse(args[1])
	globalServiceName = args[2]

	http.HandleFunc("/healthcheck", healthcheck)

	http.HandleFunc("/", serviceHandler)

	log.Printf("start [%s] bind [%s:%d]\r\n",
		globalServiceName, globalEndPoint.IP, globalEndPoint.Port)

	go RegisterEndPoint()

	err := http.ListenAndServe(globalEndPoint.Get(), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		return
	}
}
