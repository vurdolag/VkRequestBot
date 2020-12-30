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

func bot(w *sync.WaitGroup) {
	conf, _ := configs.New(*confPath)

	muGlobal := new(sync.Mutex)
	re := vk.InitRE(muGlobal)
	bd := vk.InitDataBase(muGlobal)
	answer := vk.InitAnswerDataBase(conf)
	resp := vk.InitDataResponse(muGlobal)
	accounts := vk.LoadAccount(conf)

	vkSessions := make([]*vk.VkSession, 0, len(accounts))

	for _, acc := range accounts {
		session := vk.InitVkSession(acc, re, resp, muGlobal, conf, w)
		vkSessions = append(vkSessions, session)
	}

	go server.Run(vkSessions, conf)

	p := &vk.Params{}
	p1 := &vk.Params{ToBlackList: true}
	p3 := &vk.Params{LastSeen: 3600 * 24 * 14}
	p4 := &vk.Params{
		Targets:     listClub,
		TargetGroup: "30688695",
		RndRepost:   10,
		RndLike:     10,
		RndTarget:   66,
	}

	for _, session := range vkSessions {
		a := session.Auth()
		if !a {
			continue
		}

		act := vk.InitAction(session, answer, bd)
		act.LongPool()

		act.Add(act.Online, 1, 180, p)
		act.Add(act.CheckFriends, 900, 120, p)
		act.Add(act.DelOutRequests, 900, 120, p1)
		act.Add(act.DelBadFriends, 900, 120, p3)
		act.Add(act.Reposter, 900, 120, p4)
		act.Add(act.RandomLikeFeed, 900, 120, p)

		vk.RandSleep(50, 10)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	vk.RandSleep(1, 1)
	w := new(sync.WaitGroup)
	bot(w)
	w.Wait()
}
