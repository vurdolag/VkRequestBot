package server

import (
	"VkRequestBot/configs"
	"VkRequestBot/internal/vksession"
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
	write(w, "ping")
}

func Run(vkSessions []vksession.VkSession, conf configs.ConfI) {
	http.HandleFunc("/", handler)
	_ = http.ListenAndServe(":"+conf.GetPort(), nil)
}
