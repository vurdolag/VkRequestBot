package main

import (
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


func bot () {
	rand.Seed(time.Now().UnixNano())

	muGlobal := new(sync.Mutex)
	re := InitRE(muGlobal)
	bd := InitDataBase(muGlobal)
	dataAnswer := InitAnswerDataBase()
	dataResponse := InitDataResponse(muGlobal)

	file, _ := loadFile("akks.json")
	var v []Akk
	var o []Akk
	err := json.Unmarshal(file, &v)
	if err != nil { fmt.Println("json", err) }

	for _, akk := range v {
		vk := InitVkSession(akk, re, dataResponse, muGlobal)
		vk.log("AUTH start")
		a := vk.Auth()
		vk.log("AUTH -->", a)

		if !a { continue }

		act := InitAction(vk, dataAnswer, bd)

		o = append(o, Akk{
			Name: vk.MyName,
			Id: vk.MyId,
			Login: vk.Login,
			Password: vk.Password,
			Useragent: vk.Heads,
			Proxy: vk.Proxy,
		})

		go act.LongPool()
		go act.CheckFriends()
		go act.DelOutRequests(true)
		go act.DelDogAndPornFromFriends(3600*24*14)
		go act.Reposter(listClub, "", false, 10, 10, 66, "30688695")
		go act.RandomLikeFeed()

		break
		randSleep(90, 30)
	}

	res, _ := json.Marshal(o)

	s := strings.ReplaceAll(string(res), "\",", "\",\n")

	writeNewFileTxt("newAkk.json", s)
}

func test() {
	b, _ := loadFile("temp.txt")

	j, _ := jsArr(b, "payload.[1].[0].all")

	for k, v := range j {
		fmt.Println(k, string(v))
	}

}

func main() {
	time.Sleep(time.Second*2)
	bot()
	time.Sleep(time.Second*100000000)
}

