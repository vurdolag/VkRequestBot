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
	answer := vk.InitAnswerDataBase(conf)
	accounts := vk.LoadAccount(conf)

	muGlobal := new(sync.Mutex)
	re := vk.InitRE(muGlobal)
	bd := vk.InitDataBase(muGlobal)
	resp := vk.InitDataResponse(muGlobal)

	vkSessions := make([]*vk.VkSession, 0, len(accounts))

	for _, acc := range accounts {
		session := vk.InitVkSession(acc, re, resp, muGlobal, conf, w)
		vkSessions = append(vkSessions, session)
		time.Sleep(time.Millisecond * 100)
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
		act.Add(act.CheckFriends, 1200, 60, p)
		act.Add(act.DelOutRequests, 1200, 60, p1)
		act.Add(act.DelBadFriends, 1200, 60, p3)
		act.Add(act.Reposter, 1200, 60, p4)
		act.Add(act.RandomLikeFeed, 1200, 60, p)

		vk.RandSleep(40, 1)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	vk.RandSleep(1, 1)
	w := new(sync.WaitGroup)
	bot(w)
	w.Wait()
}
