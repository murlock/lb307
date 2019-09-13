//
// Copyright 2019 Jean-Francois Smigielski
//
// This software is supplied under the terms of the MIT License, a
// copy of which should be located in the distribution where this
// file was obtained (LICENSE.txt).  A copy of the license may also be
// found online at https://opensource.org/licenses/MIT.
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

type tags struct {
	Locked bool   `json:"tag.lock"`
	Up     bool   `json:"tag.up"`
	SrvId  string `json:"tag.service_id"`
}

type Module struct {
	Address string `json:"addr"`
	Score   int    `json:"score"`
	Tags    tags   `json:"tags"`
}

type jsonError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var modules []Module

// refresh_conscience will retrieve from conscience list of server of a certain type.
// IP
func refresh_conscience(addr string, ns string, kind string) error {
	url := fmt.Sprintf("http://%s/v3.0/%s/conscience/list?type=%s", addr, ns, kind)

	fmt.Println(url)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("Error %d from conscience for %s", res.StatusCode, kind)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &modules); err != nil {
		return nil
	}

	if len(modules) == 0 {
		return fmt.Errorf("No services found for %s", kind)
	}

	// TODO(mbo): remove down, locked or instance with score 0
	for _, item := range modules {
		fmt.Println(item)
	}

	return nil
}

func lb307(rep http.ResponseWriter, req *http.Request) {
	i := rand.Intn(len(modules))
	target := modules[i]
	req.URL.Scheme = "http"
	req.URL.Host = target.Address
	loc := req.URL.String()
	rep.Header().Set("Location", loc)
	rep.WriteHeader(307)

	// this part is required for AWS
	// encoded, _ := json.Marshal(jsonError{Status: 307, Message: target.Address})
	//rep.Write(encoded)
	// log.Printf("Redirect to %s", target.Address)
}

func args() {
	flag.StringVar(&namespace, "ns", "", "Namespace")
	flag.StringVar(&conscience, "conscience", "", "Conscience IP:PORT")
	flag.StringVar(&service_type, "type", "", "Type of service to extract")
	flag.IntVar(&port, "port", 8000, "Listen port")
	flag.Parse()
}

var namespace string
var conscience string
var service_type string
var port int

func main() {
	args()
	/* TODO(mbo): refresh_conscience should be done periodically */
	err := refresh_conscience(conscience, namespace, service_type)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", lb307)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
