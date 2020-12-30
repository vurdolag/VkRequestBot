package server

import (
	"VkRequestBot/configs"
	"VkRequestBot/internal/vk"
	"fmt"
	"math/rand"
	"net/http"
	"time"
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

func Run(vkSessions []*vk.VkSession, conf configs.ConfI) {
	rand.Seed(time.Now().UnixNano())
	http.HandleFunc("/", handler)
	_ = http.ListenAndServe(":"+conf.GetPort(), nil)
}
