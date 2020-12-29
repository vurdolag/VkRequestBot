package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func InitAction (vk *VkSession, data *DataAnswer, bd *DataBase) *Action {
	act := new(Action)
	act.bd = bd
	act.vk = vk
	act.myGroupList = make([]string, 0, 150)
	b := new(BotAnswer)
	b.vk = vk
	b.bd = bd
	b.maxCountAnswer = 10
	b.data = data
	act.Bot = b
	return act
}

type Action struct {
	vk *VkSession
	Bot *BotAnswer
	bd *DataBase

	myGroupList []string
	alreadyDelDialog []int
}

func (self *Action) eventProcessing(event *Event) *Event {
	if event.comment.typeComment != "" {
		self.Bot.commentAnswer(event)
	}

	if event.messageFromChat && !IsIntInList(event.fromId, self.alreadyDelDialog) {
		self.alreadyDelDialog = append(self.alreadyDelDialog, event.fromId)
		hashDialog, err := self.vk.methods.getHashDialogs()
		if err != nil {
			self.vk.logsErr(err)
			return event
		}

		id := st(event.fromId)
		_, err = self.vk.methods.leaveChat(id, hashDialog[id][0], hashDialog[id][1])
		if err != nil {
			self.vk.logsErr(err)
		}
	}

	if event.messageFromUser {
		self.Bot.answer(event)
		if !event.empty {
			_, err := self.vk.methods.SendMessage(event.fromId, event.textOut, nil,
				event.messageId, true, true)
			if err != nil {
				self.vk.logsErr(err)
			}
		}
	}
	return event
}

func (self *Action) LongPool(){
	go self.vk.methods.LongPoll(self)
	randSleep(15, 5)
	go self.vk.methods.setOnline()
	randSleep(15, 5)
	go self.vk.methods.LongPollFeed(self)
}

func (self *Action) checkUser(userId string) (bool, error) {
	self.vk.muGlobal.Lock()
	infoUser, err := getUserInfoFromApi(userId)
	self.vk.muGlobal.Unlock()
	if infoUser == nil { return false, err }

	info := *infoUser

	if err != nil || len(info) == 0 { return false, err }

	bdAns, _ := self.bd.alreadyAddUser.get(userId)
	if userId == bdAns.Id {
		self.vk.logs(fmt.Sprintf("Юзер уже был: %d", userId))
		return false, nil
	} else {
		self.bd.alreadyAddUser.put(userId, -1)
	}

	if info[0].Deactivated != "" {
		self.vk.logs(fmt.Sprintf("Юзер заблокирован или удалён: %d", userId))
		return false, nil
	}

	if info[0].Photo_200 == "https://vk.com/images/camera_200.png?ava=1" {
		self.vk.logs(fmt.Sprintf("Без аватара: %d", userId))
		return false, nil
	}

	img, err := requestsGet(info[0].Photo_200, nil)
	if err != nil { return false, err }

	score, err := moderationImg(img)
	if err != nil || len(score) != 4 { return false, err }

	if score["adult"] > 0.9 || score["gruesome"] > 0.9 {
		self.vk.logs(fmt.Sprintf("Неприемлимый аватар: %d", userId))
		return false, nil
	}

	return true, nil
}

func (self *Action) acceptOrDeclineNewFriend() error {
	randSleep(15, 5)

	self.vk.logs(fm("Запуск проверки новых друзей"))

	newFriend, err := self.vk.methods.getNewFriendList()
	if err != nil { self.vk.logs(fm("ошибка проверки новых друзей")); return err }

	if len(newFriend) > 0 {
		self.vk.logs(fm("Новых друзей найдено: %d", len(newFriend)))
	} else {
		self.vk.logs(fm("Нет новых друзей..."))
		return nil
	}

	add, notAdd := 0, 0

	for userId, listHash := range newFriend {
		if len(listHash) != 2 { continue }

		check, _ := self.checkUser(userId)
		id, _ := strconv.Atoi(userId)

		if id == 0 { continue }

		if check {
			add++
			_ = self.vk.methods.friendAccept(id, listHash[0])
		} else {
			notAdd++
			_ = self.vk.methods.friendDecline(id, listHash[1])
		}
		randSleep(45, 15)
	}

	self.vk.logs(fm("Друзей добавил: %d из %d", add, add+notAdd))

	return nil
}

func (self *Action) CheckFriends() {
	randSleep(600, 180)
	for self.vk.methods.working {
		err := self.acceptOrDeclineNewFriend()
		if err != nil {
			self.vk.logsErr(err)
		}
		randSleep(3600, 1800)
	}
}

func (self *Action) DelOutRequests(toBlackList bool) {
	randSleep(600, 180)
	for self.vk.methods.working {
		err := self.vk.methods.DelOutRequests(toBlackList)
		if err != nil { self.vk.logsErr(err) }
		randSleep(7200, 3600)
	}
}

func (self *Action) Reposter(targets []string, msg string, fromGroup bool,
		                     rndRepost, rndLike, rndTarget float32, targetGroup string) {

	randSleep(600, 180)
	for self.vk.methods.working {
		self.vk.log("reposter start")
		likeFrom := "feed_recent"
		groupId := ""
		if fromGroup {
			likeFrom = "wall_page"
			if len(self.myGroupList) == 0 {
				groups, err := self.vk.methods.getGroups(self.vk.MyId)
				if err != nil {
					self.vk.logsErr(err)
					randSleep(600, 300)
					continue
				}

				for _, i := range *groups {
					self.myGroupList = append(self.myGroupList, i.Id)
				}
			}
			l := len(self.myGroupList)
			groupId = fm("club%s", self.myGroupList[rand.Intn(l)][1:])
			targets = []string{ groupId }
		}

		//club53454
		target := targets[rand.Intn(len(targets))]
		postIds, err := self.vk.methods.getPostFrom(target, nil, rand.Intn(10) + 10)
		if err != nil {
			self.vk.logsErr(err)
			continue
		}
		if len(postIds) != 0 {
			postID := postIds[0][0]
			post := fm("%s_%s", self.vk.MyId, postID)

			newRepost, _ := self.bd.newRepost.get(post)

			if post != newRepost.Id {
				_, err = self.bd.newRepost.put(post, -1)
				if err != nil { self.vk.logsErr(err) }

				if isRand(90 ) {
					t := make([]string, 0, 20)
					for _, v := range postIds { t = append(t, v[0]) }
					self.vk.methods.viewPost(t, []string{}, 40)
				}

				if isRand(rndRepost) {
					self.vk.methods.Repost(postID, msg)
				} else {
					if isRand(rndLike) || targetGroup != "" && (strings.Contains(postID, groupId) && isRand(rndTarget)) {
						self.vk.methods.Like(postID, postIds[0][1], likeFrom)
					} else {
						self.vk.logs("reposter loss...")
					}
				}
			}
		}
		randSleep(3200, 1200)
	}
}

func (self *Action) RandomLikeFeed() {
	randSleep(600, 180)
	for self.vk.methods.working {
		self.vk.log("random like feed start")

		likeFeed, err := self.vk.methods.getPostFrom("", nil, rand.Intn(9) + 7)
		if err != nil || len(likeFeed) == 0 {
			self.vk.logsErr(err)
			randSleep(1200, 300)
			continue
		}

		randSleep(20, 10)

		post := likeFeed[rand.Intn(len(likeFeed))]

		if len(post) != 2 {
			self.vk.log("error post len != 2")
			randSleep(1200, 300)
			continue
		}

		postId := post[0]
		hashLike := post[1]

		postIdForView := make([]string, 0, 15)
		for i := range likeFeed { postIdForView = append(postIdForView, likeFeed[i][0]) }

		self.vk.methods.viewPost(postIdForView, []string{}, 40)

		randSleep(25, 10)

		self.vk.methods.Like(postId, hashLike, "feed_recent")

		randSleep(3200, 1800)
	}
}

func (self *Action) DelDogAndPornFromFriends (lastSeen int) {
	randSleep(600, 180)
	for self.vk.methods.working {
		self.vk.logs("Запуск удаления собак и непристойных юзеров")

		friendList, err := self.vk.methods.getFriends(self.vk.MyId)
		if err != nil {
			self.vk.logsErr(err)
			randSleep(900, 600)
			continue
		}

		friend := *friendList
		lenFriend := len(friend)

		temp := make([]string, lenFriend)
		randSlice := rand.Perm(lenFriend-1)
		for i, j := range randSlice {
			temp[i] = friend[j].Id
		}

		self.vk.muGlobal.Lock()
		infoUser, er := getUserInfoFromApi(temp...)
		if er != nil { self.vk.logsErr(er) }
		self.vk.muGlobal.Unlock()

		info := *infoUser

		self.vk.logs(fm("Найдено друзей: %d (%d)", len(info), lenFriend))

		count := 0
		for index := range info {
			flag := 0

			if info[index].Photo_200 == "https://vk.com/images/camera_200.png?ava=1" {
				flag = 1
			}

			if info[index].Deactivated != "" { flag = 1}

			if lastSeen != -1 && int(time.Now().Unix()) - info[index].Last_seen.Time > lastSeen { flag = 1}

			if flag == 1 {
				count++
				_, _ = self.vk.methods.delFriend(info[index].Id, false)
				id := st(info[index].Id)
				bdAns, _ := self.bd.alreadyAddUser.get(id)
				if bdAns.Id != id {
					randSleep(1, 0)
					_, _ = self.bd.alreadyAddUser.put(id, -1)
				}
			}
		}
		self.vk.logs(fm("Удалено друзей: %d из %d", count, len(info)))

		randSleep(7200, 3600)
	}
}



func InitAnswerDataBase() *DataAnswer {
	var dataList = make([][][]string, 0, 700)
	var endDataList = make([]string, 0, 30)

	data, err := loadFile("answer/text.txt")
	if err != nil {panic(err)}

	for _, vals := range strings.Split(string(data), "\n") {
		if vals == "" { continue }
		temps := make([][]string, 0, 10)
		for index, val := range strings.Split(vals, "||"){
			temp := make([]string, 0, 10)
			for _, v := range strings.Split(val, ",,") {
				if index == 0 { v = enc(v) }
				temp = append(temp, v)
			}
			if index == 0 {
				if temp[0] == "#" {
					temp = []string{fmt.Sprintf("(%s)", strings.Join(temp[1:], "|"))}
					} else {
						temp = []string{fmt.Sprintf(`(\b%s\b)`, strings.Join(temp, `\b|\b`))}
					 	}
					}
			temps = append(temps, temp)
		}

		if len(temps) != 2 { continue }
		dataList = append(dataList, temps)
	}

	endData, err := loadFile("answer/EndPhrase.txt")
	if err != nil {panic(err)}

	endDataList = strings.Split(string(endData), ",,")

	return &DataAnswer{
		botDataBase: dataList,
		endAnswerList: endDataList,
	}
}


type DataAnswer struct {
	botDataBase [][][]string
	endAnswerList []string
}

type BotAnswer struct {
	vk *VkSession
	maxCountAnswer int
	data *DataAnswer
	bd *DataBase
}

func (self *BotAnswer) answer(event *Event) *Event {
	maxAnswer := self.preparation(event)

	if maxAnswer > self.maxCountAnswer {
		self.vk.log(event.fromId, maxAnswer, "MAX <-", event.text)
		event.empty = true
		return event
	}

	//if event.attachment

	//message_text, english = await self.is_english(message_text, event.from_id)

	targetAnswer, _ := self.getAnswer(event)

	targetAnswer = self.insertName(st(event.fromId), targetAnswer)

	targetAnswer = self.getEndAnswer(targetAnswer, maxAnswer)

	//if english:
	//# проверяет правильность исходящего сообщения для правильного перевода
	//target_answer = await Utils.checker_text(target_answer)
	//# перевод
	//target_answer = await Utils.translate(target_answer, 'ru-en')

	event.Answer(targetAnswer)

	self.log(event, maxAnswer)

	return event
}

func (self *BotAnswer) commentAnswer(event *Event) *Event {
	self.answer(event)

	if event.textOut == "" { return event }

	idPost := event.comment.idPost
	fromId := event.comment.fromId
	reply := event.comment.reply

	randSleep(240, 120)

	if event.comment.typeComment == "post_reply" {
		_ = self.vk.methods.commentPost(idPost, event.textOut, fromId, reply, "")
	}

	if event.comment.typeComment == "comment_photo" {
		_ = self.vk.methods.commentPhoto(idPost, event.textOut, "", fromId)
	}
	return event
}

func (self *BotAnswer) voiceMessageProcessing(event Event, messageText string, maxAnswer int) string {
	//var voiceText string
	return ""
}

func (self *BotAnswer) getAnswer(event *Event) (string, [][][]string) {
	message := enc(strings.ToLower(event.text))

	t1 := time.Now().UnixNano()

	var answers = make([][][]string, 0, 10)
	for index := range self.data.botDataBase {
		if self.vk.re.Check(message, self.data.botDataBase[index][0][0]) {
			answers = append(answers, self.data.botDataBase[index])
		}
	}

	fmt.Println("time ->", time.Now().UnixNano() - t1)

	if len(answers) > 0 {
		return randChoice(answers[0][1]), answers
	} else { return "", answers }
}

func (self *BotAnswer) getEndAnswer(targetAnswer string, maxAnswer int) string {
	if maxAnswer == self.maxCountAnswer {
		targetAnswer = randChoice(self.data.endAnswerList)
	}
	return targetAnswer
}

func (self *BotAnswer) isEnglish(event *Event) {
	//pass
}

func (self *BotAnswer) insertName(userId, messageText string) string {
	if strings.Contains(messageText, "*fname*") {
		name, err := getUserInfoFromApi(userId)
		if err != nil {
			messageText = strings.ReplaceAll(messageText, "*fname*", "")
			return messageText
		}

		n := *name
		messageText = strings.ReplaceAll(messageText, "*fname*", n[0].First_name)
	}
	return messageText
}

func (self *BotAnswer) preparation(event *Event) int {
	qBd := fm("%s_%d", self.vk.MyId, event.fromId)
	maxAnswer, err := self.bd.maxAnswer.get(qBd)

	if err == nil && maxAnswer.Id == qBd {
		_, _ = self.bd.maxAnswer.up(qBd, maxAnswer.Count + 1)
		return maxAnswer.Count
	} else {
		_, _ = self.bd.maxAnswer.put(qBd, 1)
		return 0
	}
}

func (self *BotAnswer) log(event *Event, maxAnswer int) {
	self.vk.log(event.fromId, maxAnswer, "<-", event.text)
	self.vk.logs(fm("%d, %d <- %s", event.fromId, maxAnswer, event.text), "msg/" + self.vk.Login)

	if event.textOut == "" {
		self.vk.logs(fm("%d, %d <- %s", event.fromId, maxAnswer, event.text), "msg_not_answer")
	} else {
		self.vk.logs(fm("%d, %d -> %s", event.fromId, maxAnswer, event.textOut), "msg/" + self.vk.Login)
		self.vk.log(event.fromId, maxAnswer, "->", event.textOut)
	}
}