package main

import (
	vksession "VkRequestBot/internal"
	server "VkRequestBot/internal/server"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

var listClub = []string{
	"public193767860",
	"club30688695",
	"club168691465",
	"club178894882",
	"club193674464",
	"club174587092",
}

func bot() {
	rand.Seed(time.Now().UnixNano())

	muGlobal := new(sync.Mutex)
	re := vksession.InitRE(muGlobal)
	bd := vksession.InitDataBase(muGlobal)
	dataAnswer := vksession.InitAnswerDataBase()
	dataResponse := vksession.InitDataResponse(muGlobal)

	file, _ := vksession.LoadFile("akks.json")
	var v []vksession.Akk
	var o []vksession.Akk
	err := json.Unmarshal(file, &v)
	if err != nil {
		fmt.Println("json", err)
	}

	for _, akk := range v {
		vk := vksession.InitVkSession(akk, re, dataResponse, muGlobal)
		a := vk.Auth()

		if !a {
			continue
		}

		act := vksession.InitAction(vk, dataAnswer, bd)

		o = append(o, vksession.Akk{
			Name:      vk.MyName,
			Id:        vk.MyId,
			Login:     vk.Login,
			Password:  vk.Password,
			Useragent: vk.Heads,
			Proxy:     vk.Proxy,
		})

		go act.LongPool()
		go act.CheckFriends()
		go act.DelOutRequests(true)
		go act.DelDogAndPornFromFriends(3600 * 24 * 14)
		go act.Reposter(listClub, "", false, 10, 10, 66, "30688695")
		go act.RandomLikeFeed()

		vksession.RandSleep(90, 30)
	}

	res, _ := json.Marshal(o)

	s := strings.ReplaceAll(string(res), "\",", "\",\n")

	vksession.WriteFile("newAkk.json", s)

	go server.Run()
}

/*
func test() {
	b, _ := vksession.LoadFile("temp.txt")

	j, _ := vksession.jsArr(b, "payload.[1].[0].all")

	for k, v := range j {
		fmt.Println(k, string(v))
	}

}

*/

func main() {
	time.Sleep(time.Second * 2)
	bot()
	time.Sleep(time.Second * 100000000)
}
