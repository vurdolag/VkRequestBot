package main

import (
	"VkRequestBot/configs"
	"VkRequestBot/internal/server"
	"VkRequestBot/internal/vk"
	"flag"
	"math/rand"
	"sync"
	"time"
)

var confPath = flag.String("conf-path", "./configs/.env", "path to conf .env")

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

	conf, _ := configs.New(*confPath)

	muGlobal := new(sync.Mutex)
	re := vk.InitRE(muGlobal)
	bd := vk.InitDataBase(muGlobal)
	dataAnswer := vk.InitAnswerDataBase(conf)
	dataResponse := vk.InitDataResponse(muGlobal)
	accounts := vk.LoadAccount(conf)

	vkSessions := make([]*vk.VkSession, 0, len(accounts))

	for _, acc := range accounts {
		v := vk.InitVkSession(acc, re, dataResponse, muGlobal, conf)
		vkSessions = append(vkSessions, v)
	}

	go server.Run(vkSessions, conf)

	for _, v := range vkSessions {
		vk.RandSleep(10, 3)
		a := v.Auth()
		if !a {
			continue
		}

		act := vk.InitAction(v, dataAnswer, bd)

		go act.LongPool()
		go act.CheckFriends()
		go act.DelOutRequests(true)
		go act.DelDogAndPornFromFriends(3600 * 24 * 14)
		go act.Reposter(listClub, "", false, 10, 10, 66, "30688695")
		go act.RandomLikeFeed()

		vk.RandSleep(90, 30)
		break
	}

}

func main() {
	time.Sleep(time.Second * 2)
	bot()
	time.Sleep(time.Second * 100000000)
}
