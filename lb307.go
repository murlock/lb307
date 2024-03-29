//
// Copyright 2019 Jean-Francois Smigielski
//
// This software is supplied under the terms of the MIT License, a
// copy of which should be located in the distribution where this
// file was obtained (LICENSE.txt).  A copy of the license may also be
// found online at https://opensource.org/licenses/MIT.
//

package main

import(
	"encoding/json"
	"strings"
	"flag"
	"log"
	"os"
	"bufio"
	"net/http"
	"math/rand"
)

type jsonError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var targets []string

func lb307(rep http.ResponseWriter, req *http.Request) {
	i := rand.Intn(len(targets))
	target := targets[i]
	encoded, _ := json.Marshal(jsonError{Status: 307, Message: target})
	req.URL.Scheme = "http"
	req.URL.Host = target
	loc := req.URL.String()
	rep.Header().Set("Location", loc)
	rep.WriteHeader(307)
	rep.Write(encoded)
}

func main() {
	targets = make([]string, 0)

	flag.Parse()
	for i:=0; i<flag.NArg(); i++ {
		path := flag.Arg(i)
		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		} else {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				t := scanner.Text()
				t = strings.TrimLeft(t, " \t\r\n")
				t = strings.TrimRight(t, " \t\r\n")
				if t != "" && ! strings.HasPrefix(t, "#") {
					targets = append(targets, t)
				}
			}
			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
			f.Close()
		}
	}

	if len(targets) <= 0 {
		log.Fatal("No target")
	}

	http.HandleFunc("/", lb307)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
