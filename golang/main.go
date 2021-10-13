package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func getRemote(url string, results chan<- string) {
	resp, err := http.Get(url)
	if err != nil {
		results <- err.Error()
		return
	}
	// do not forget to close body
	defer resp.Body.Close()
	body, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		results <- err2.Error()
	} else {
		results <- string(body)
	}
}

func main() {
	http.HandleFunc("/api/root-service", func(w http.ResponseWriter, r *http.Request) {
		nodeResult := make(chan string)
		pythonResult := make(chan string)
		// Parallel run
		go getRemote("http://node-service/api/service1", nodeResult)
		go getRemote("http://python-service/", pythonResult)

		// Thanks to channels code will be executed sequently
		response1Message := fmt.Sprintf("node response: %s \n\n", <-nodeResult)
		response2Message := fmt.Sprintf("python response: %s \n\n", <-pythonResult)

		fmt.Fprintf(w, response1Message+response2Message)
	})
	http.ListenAndServe(":8080", nil)
}
