package vksession

import (
	"fmt"
	"regexp"
	"strconv"
)

func InitComment(update FeedUpdate, re *RE) Comment {
	c := Comment{}
	c.pars(update, re)
	return c
}

type Comment struct {
	idPost      string
	reply       string
	thread      string
	link        string
	typeComment string
	fromId      int
}

func (self *Comment) pars(up FeedUpdate, re *RE) {
	update := up.Link + "&"

	reply := re.FindAllBetweenStr(update, "reply=", "&")
	if len(reply) != 0 {
		self.reply = reply[0]
	}

	thread := re.FindAllBetweenStr(update, "thread=", "&")
	if len(thread) != 0 {
		self.thread = thread[0]
	}

	idPost := re.FindAllBetweenStr(update, "com/", "\\?")
	if len(idPost) != 0 {
		self.idPost = idPost[0]
	} else {
		idPost = re.FindAllBetweenStr(update, "com/", "&")
		if len(idPost) != 0 {
			self.idPost = idPost[0]
		}
	}

	fromId := re.FindAllBetweenStr(up.Title, "mention_id=", "\"")
	if len(fromId) != 0 && len(fromId[0]) > 2 {
		id, err := strconv.Atoi(fromId[0][3:])
		if err == nil {
			fmt.Println(err)
			self.fromId = id
		}
	}
}

var reEventPars1, _ = regexp.Compile("<b.*&b>")
var reEventPars2, _ = regexp.Compile("<span.*/span>")
var reEventPars3, _ = regexp.Compile("&quot;")
var reEventPars4, _ = regexp.Compile("<.*?>")
var reEventPars5, _ = regexp.Compile("<.*?=")

var eventUpdateTypeId = []int{33, 49, 2097185, 1, 17}

type Event struct {
	empty           bool
	id              string
	text            string
	textOut         string
	update          []interface{}
	messageId       int
	fromId          int
	attachment      Atta
	timeStamp       int
	messageFromUser bool
	messageFromChat bool
	fromFeed        bool
	comment         Comment
}

func (self *Event) pars(update []byte, fromFeed bool, re *RE) *Event {
	if fromFeed {
		j := NewJs(update)

		text := finderFirst(update, "\"text\":\"", "\",\"")
		title := finderFirst(update, "\"title\":", "\",\"text\":")

		feed := FeedUpdate{
			Version:   j.Int("version"),
			Type:      j.Get("type"),
			Link:      j.Get("link"),
			Text:      text,
			Author_id: j.Int("author_id"),
			Title:     title,
		}

		self.fromFeed = true

		if feed.Type == "post_reply" || feed.Type == "comment_photo" {
			self.text = self.clearMessage(feed.Text)

			if self.text == "" {
				self.empty = true
				return self
			}
			comment := InitComment(feed, re)
			comment.typeComment = feed.Type
			self.comment = comment
			self.fromId = comment.fromId
			return self
		}
	}

	if update != nil && !fromFeed {
		j := NewJsArray(update)
		typeUp := j.Int(0)
		typeUpId := j.Int(2)
		fromId := j.Int(3)
		if j.len > 6 && typeUp == 4 {
			//[4, 692687, 33, 589347890, 1598359476, ' ... ', 'Приехать можешь м медветково?', {}, 0]
			if typeUpId > 100 && fromId > 2000000000 {
				self.messageFromChat = true
				self.fromId = fromId
				return self
			}

			if IsIntInList(typeUpId, eventUpdateTypeId) {
				self.messageFromUser = true
				self.text = j.Get(5)
				//self.attachment = nil
				self.fromId = fromId
				self.timeStamp = j.Int(4)
				self.messageId = j.Int(1)
				return self
			}
		}
	}
	return self
}

func (self *Event) clearMessage(str string) string {
	if len(reEventPars1.FindAllString(str, -1)) != 0 {
		return ""
	}
	str = reEventPars3.ReplaceAllString(str, "\"")
	str = reEventPars2.ReplaceAllString(str, "")
	str = reEventPars4.ReplaceAllString(str, "")
	return reEventPars5.ReplaceAllString(str, "")
}

func (self *Event) Answer(msg string) *Event {
	self.textOut = msg
	return self

}
