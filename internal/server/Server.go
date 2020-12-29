package server

import (
	"fmt"
	"net/http"
)

func write(w http.ResponseWriter, msg string) {
	_, err := w.Write([]byte(msg))
	if err != nil {
		fmt.Println(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {

}

func Run() {
	http.HandleFunc("/", handler)
	_ = http.ListenAndServe(":6060", nil)
}
