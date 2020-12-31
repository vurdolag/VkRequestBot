package vk

import (
	"VkRequestBot/configs"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html/charset"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var reSubBackSlash, _ = regexp.Compile(`\\`)
var reFindSubHash, _ = regexp.Compile("[a-z0-9_-]+")
var reFigureScope, _ = regexp.Compile(`{.+?}`)
var rePostId, _ = regexp.Compile(`wall-*?\d+?_\d+`)
var reFindD, _ = regexp.Compile("[0-9]+")
var reD, _ = regexp.Compile(`[^0-9]`)
var reSubS, _ = regexp.Compile("[^0-9-_]")

var CHECKOK = []byte("{\"payload\":[0")

type Response struct {
	id     int
	data   *DataResponse
	start  int
	end    int
	vk     *VkSession
	params map[string]string
	url    string
	check  []byte
}

func (self *Response) Finder(str1, str2 string, count int) []string {
	r := make([]string, 0, 15)
	var a, y, end, index, c int

	s1 := []byte(str1)
	s2 := []byte(str2)

	self.data.muGlobal.Lock()
	pos := self.data.get(self.id)
	index = pos[0]
	end = pos[1]

	a = -1
	switcher := true

	for index <= end {
		if y == 0 && self.data.data[index] == s1[0] && switcher {
			//fmt.Println(string(self.data.data[index]), string(s1[0]))
			if len(s1) == 1 {
				a = index
				switcher = false
				index++
			} else {
				for x := 1; x < len(s1); x++ {
					index++
					//fmt.Println(string(self.data.data[index]), string(s1[x]))
					if self.data.data[index] != s1[x] {
						a = -1
						break
					}
					if x == len(s1)-1 {
						a = index
						switcher = false
						index++
					}
				}
			}
		}
		if a != -1 && self.data.data[index] == s2[0] {
			//fmt.Println(string(self.data.data[index]), string(s2[0]))
			if len(s2) == 1 {
				y = index - 1
				switcher = true
			} else {
				for x := 1; x < len(s2); x++ {
					index++
					//fmt.Println(string(self.data.data[index]), string(s2[x]))
					if self.data.data[index] != s2[x] {
						y = 0
						break
					}
					if x == len(s2)-1 {
						y = index - len(s2)
						switcher = true
					}
				}
			}
		}
		if y != 0 {
			r = append(r, string(self.data.data[a+1:y+1]))
			a, y = -1, 0
			c++
			if c == count {
				break
			}
		}
		index++
	}
	self.data.muGlobal.Unlock()
	return r
}

func (self *Response) Find(str1, str2 string) []string {
	return self.Finder(str1, str2, -1)
}

func (self *Response) FindFirst(str1, str2 string) string {
	v := self.Finder(str1, str2, 1)
	if len(v) != 0 {
		return v[0]
	} else {
		return ""
	}
}

func (self *Response) js(q ...string) (string, error) {
	i, err := js(self.getByte(-1), q...)
	return i, err
}

func (self *Response) jsAsByte(q ...string) ([]byte, error) {
	i, err := jsAsByte(self.getByte(-1), q...)
	return i, err
}

func (self *Response) jsInt(q ...string) (int, error) {
	i, err := jsInt(self.getByte(-1), q...)
	return i, err
}

func (self *Response) jsStr(q ...string) (string, error) {
	i, err := jsStr(self.getByte(-1), q...)
	return i, err
}

func (self *Response) jsArr(q ...string) ([][]byte, error) {
	i, err := jsArr(self.getByte(-1), q...)
	return i, err
}

func (self *Response) NewJsArray(q ...string) *JsArray {
	b, err := self.jsAsByte(q...)
	if err != nil {
		logsErr(err)
	}
	return NewJsArray(b)
}

func (self *Response) NewJs(q ...string) *Js {
	b, err := self.jsAsByte(q...)
	if err != nil {
		logsErr(err)
	}
	return NewJs(b)
}

func (self *Response) Check(msg string, err error) bool {
	if self.vk == nil {
		return false
	}

	if err != nil {
		self.vk.log("ERROR! -->", msg, "-->", err)
		self.vk.logs(fm("ERROR! msg = %s err = %v", msg, err), "error")
		return false
	}

	if len(self.check) > 0 && self.check[len(self.check)-1] != []byte("0")[0] {
		self.vk.log("ERROR! -->", msg)
		self.vk.logs(fm("ERROR! msg = %s\n\turl = %s\n\tparams = %v\n\tbody: %s\n",
			msg, self.url, self.params, string(self.getByte(100))), "vk_error")
		return false
	}
	return true
}

func (self *Response) In(str string) bool {
	s1 := []byte(str)

	self.data.muGlobal.Lock()
	pos := self.data.get(self.id)
	index := pos[0]
	end := pos[1]

	for index <= end {
		if self.data.data[index] == s1[0] {
			for x := 1; x < len(s1); x++ {
				index++
				if self.data.data[index] != s1[x] {
					break
				}
				if x == len(s1)-1 {
					self.data.muGlobal.Unlock()
					return true
				}
			}
		}
		index++
	}
	self.data.muGlobal.Unlock()
	return false
}

func (self *Response) Close() {
	self.data.muGlobal.Lock()
	self.data.del(self.id)
	self.data.muGlobal.Unlock()
}

func (self *Response) getByte(count int) []byte {
	self.data.muGlobal.Lock()
	pos := self.data.get(self.id)
	index := pos[0]
	end := pos[1]
	temp := make([]byte, 0, end-index+10)
	c := 0
	for index < pos[1] {
		temp = append(temp, self.data.data[index])
		index++
		c++
		if c == count {
			break
		}
	}
	self.data.muGlobal.Unlock()
	return temp
}

func (self *Response) Str() string {
	return string(self.getByte(-1))
}

type DataResponse struct {
	responseIds    int
	data           []byte
	lastWriteIndex int
	activeResponse [][]int
	muGlobal       *sync.Mutex
}

func InitDataResponse(muGlobal *sync.Mutex) *DataResponse {
	r := &DataResponse{
		data:           make([]byte, 10000000, 10000000),
		responseIds:    0,
		lastWriteIndex: 0,
		activeResponse: make([][]int, 120, 120),
		muGlobal:       muGlobal,
	}

	for i := range r.activeResponse {
		r.activeResponse[i] = []int{0, len(r.data) + 1, 0}
	}

	return r
}

func (self *DataResponse) test_t() int {
	ind := 0
	for _, val := range self.activeResponse {
		if val[2] != 0 {
			ind++
		}
	}
	return ind
}

func (self *DataResponse) Write(b []byte) *Response {
	self.muGlobal.Lock()

	self.responseIds++

	key := self.responseIds
	index := self.lastWriteIndex
	start := index
	end := start + len(b)

	if len(self.data) < index+len(b) {
		logs(fm("optim start -> b = %d, buf = %d, lP = %d, lA = %d", len(b), len(self.data), self.lastWriteIndex, self.test_t()), "writer")
		self.Optimise()
		logs(fm("optim end   -> b = %d, buf = %d, lP = %d, lA = %d", len(b), len(self.data), self.lastWriteIndex, self.test_t()), "writer")
		index = self.lastWriteIndex
		start = index
		end = start + len(b)
	}

	if len(self.data) < index+len(b)+1 {
		logs(fm("alloc start -> b = %d, buf = %d, lP = %d, lA = %d", len(b), len(self.data), self.lastWriteIndex, self.test_t()), "writer")
		data := make([]byte, len(self.data)*2, cap(self.data)*2)
		for x := range self.data {
			data = append(data, self.data[x])
		}
		self.data = data
		logs(fm("alloc end   -> b = %d, buf = %d, lP = %d, lA = %d", len(b), len(self.data), self.lastWriteIndex, self.test_t()), "writer")
	}

	self.add(start, end, key)
	self.lastWriteIndex = end

	check := make([]byte, 0, len(CHECKOK))

	for i := range b {
		self.data[index] = b[i]
		index++
		if i < len(CHECKOK) {
			check = append(check, b[i])
		}
	}

	response := &Response{
		id:    key,
		data:  self,
		start: start,
		end:   end,
		check: check,
	}

	self.muGlobal.Unlock()

	return response
}

func (self *DataResponse) WriteStr(s string) *Response {
	return self.Write([]byte(s))
}

func (self *DataResponse) Optimise() {
	var startP, endP, key, count, lastMaxPos int

	self.sorter()
	for index, x := range self.activeResponse {
		startP = x[0]
		endP = x[1]
		key = x[2]

		if key == 0 {
			continue
		}

		if lastMaxPos != startP {
			count = startP - lastMaxPos
			self.move(lastMaxPos, count, endP-startP)
		}

		startP -= count
		endP -= count

		self.activeResponse[index] = []int{startP, endP, key}

		lastMaxPos = endP
	}
	self.lastWriteIndex = lastMaxPos
	runtime.GC()
}

func (self *DataResponse) move(start, count, l int) {
	var i, j, index int
	index = start
	for index < start+count+l {
		i = index + count
		j = index
		if i > len(self.data)-1 {
			i = 0
		}
		if j > len(self.data)-1 {
			j = len(self.data) - 1
		}
		self.data[j] = self.data[i]
		index++
	}
}

func (self *DataResponse) sorter() {
	sort.SliceStable(self.activeResponse, func(i, j int) bool {
		return self.activeResponse[i][1] < self.activeResponse[j][1]
	})

}

func (self *DataResponse) add(start, end, key int) bool {
	self.sorter()
	for index, val := range self.activeResponse {
		if val[2] == 0 {
			self.activeResponse[index] = []int{start, end, key}
			return true
		}
	}
	temp := make([][]int, len(self.activeResponse)*2, cap(self.activeResponse)*2)
	for i := range temp {
		if i < len(self.activeResponse) {
			temp[i] = self.activeResponse[i]
		} else {
			temp[i] = []int{0, len(self.data) + 1, 0}
		}
	}
	self.activeResponse = temp
	self.add(start, end, key)
	return true
}

func (self *DataResponse) get(key int) []int {
	for index, val := range self.activeResponse {
		if val[2] == key {
			return self.activeResponse[index]
		}
	}
	return []int{0, len(self.data) + 1, 0}
}

func (self *DataResponse) del(key int) {
	if key != 0 {
		for index, val := range self.activeResponse {
			if val[2] == key {
				self.activeResponse[index] = []int{0, len(self.data) + 1, 0}
				break
			}
		}
	}
}

func (self *DataResponse) New(vk *VkSession, params map[string]string, url string) *Response {
	return &Response{
		id:     0,
		data:   self,
		start:  0,
		end:    0,
		vk:     vk,
		params: params,
		url:    url,
	}
}

type VkSession struct {
	Session  *http.Client
	Heads    string
	Proxy    string
	Login    string
	Password string
	MyId     string
	MyName   string
	muLocal  *sync.Mutex
	muGlobal *sync.Mutex

	re       *RE
	methods  *Methods
	response *DataResponse

	conf configs.ConfI
	wait *sync.WaitGroup
}

func InitVkSession(akk Akk, re *RE, resp *DataResponse, mu *sync.Mutex,
	conf configs.ConfI, w *sync.WaitGroup) *VkSession {
	rand.Seed(time.Now().UnixNano())
	w.Add(1)
	methods := InitMethods()
	vk := &VkSession{
		Login: akk.Login, Password: akk.Password,
		Proxy:    fmt.Sprintf("http://%s", akk.Proxy),
		Heads:    akk.Useragent,
		re:       re,
		muGlobal: mu,
		muLocal:  new(sync.Mutex),
		methods:  methods,
		response: resp,
		conf:     conf,
		wait:     w,
	}
	methods.vk = vk
	return vk
}

func (self *VkSession) Auth() bool {
	jar, _ := cookiejar.New(nil)

	proxyUrl, _ := url.Parse(self.Proxy)
	cookie, err := self.getCookies()

	if err != nil {
		self.log(err)
		self.enterVk()
	} else {
		u, _ := url.Parse("http://vk.com")
		jar.SetCookies(u, cookie)
	}

	self.Session = &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		Timeout:   time.Second * 60,
		Jar:       jar,
	}

	if self.check() {
		self.logs("Авторизация OK")
		self.log("AUTH OK!")
		return true
	} else {
		self.logs("Авторизация ОШИБКА")
		self.log("AUTH FALSE!")
		return false
	}
}

func (self *VkSession) getCookies() ([]*http.Cookie, error) {
	file, err := loadFile(self.conf.GetCookiePath() + self.Login + ".json")
	if err != nil {
		return nil, err
	}
	var cookies []*http.Cookie
	var j map[string]interface{}
	err = json.Unmarshal(file, &j)
	if err != nil {
		return nil, err
	}

	if _, ok := j["time_ex"]; ok {
		delete(j, "time_ex")
	}

	for key, val := range j {
		cook := &http.Cookie{
			Name:   key,
			Value:  val.(string),
			Path:   "/",
			Domain: "vk.com",
		}
		cookies = append(cookies, cook)
	}

	return cookies, nil
}

func (self *VkSession) saveCookies() {
	//TODO
}

func (self *VkSession) getHeaders(typeReq string) map[string]string {
	head := map[string]string{
		"User-Agent":      self.Heads,
		"Accept":          "*/*",
		"Accept-Language": "en-US,en;q=0.9,ru;q=0.8,ja;q=0.7",
		//"Cookie": self.CookiesString,
		//"Accept-Encoding": "gzip, deflate, br",
		"Connection": "keep-alive",
		"DNT":        "1",
	}

	if typeReq == "POST" {
		head["content-type"] = "application/x-www-form-urlencoded"
		head["x-requested-with"] = "XMLHttpRequest"
	}

	return head

}

func (self *VkSession) enterVk() bool {
	//TODO
	return true
}

func (self *VkSession) GET(targetUrl string) (*Response, error) {
	r := self.response.New(self, nil, targetUrl)
	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return r, err
	}

	for key, val := range self.getHeaders("GET") {
		req.Header.Set(key, val)
	}

	resp, err := self.Session.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	respBody, err := charset.NewReader(
		resp.Body,
		resp.Header.Get("Content-Type"))
	if err != nil {
		return r, err
	}

	response, err := ioutil.ReadAll(respBody)
	if err != nil {
		return r, err
	}

	r = self.response.Write(response)
	r.vk = self
	r.params = nil
	r.url = targetUrl
	return r, nil
}

func (self *VkSession) POST(targetUrl string, params map[string]string) (*Response, error) {
	r := self.response.New(self, params, targetUrl)
	postData := url.Values{}
	for key, val := range params {
		postData.Set(key, val)
	}

	req, err := http.NewRequest("POST", targetUrl, strings.NewReader(postData.Encode()))
	if err != nil {
		return r, err
	}

	for key, val := range self.getHeaders("POST") {
		req.Header.Set(key, val)
	}

	resp, err := self.Session.Do(req)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()

	respBody, err := charset.NewReader(
		resp.Body,
		resp.Header.Get("Content-Type"))
	if err != nil {
		return r, err
	}

	res, err := ioutil.ReadAll(respBody)
	if err != nil {
		return r, err
	}

	r = self.response.Write(res)
	r.vk = self
	r.params = params
	r.url = targetUrl
	return r, nil
}

func (self *VkSession) check() bool {
	res, err := self.GET("https://vk.com/feed")
	defer res.Close()
	if err != nil {
		self.logsErr(err)
		return false
	}

	if !self.methods.checkStatus(res) {
		return false
	}

	id := res.FindFirst("id: ", ",")
	if id == "" || id == "0" {
		return self.enterVk()
	} else {
		self.MyId = id
	}

	self.MyName = res.FindFirst("\"top_profile_name\">", "</div>")

	return true
}

func (self *VkSession) log(msg ...interface{}) {
	t := time.Now().Unix()
	v := fm("%v", msg)
	f := fm("%d > %12.12s %10.10s %9.9s > %s", t, self.Login, self.MyName, self.MyId, v[1:len(v)-1])
	fmt.Println(f)
}

func (self *VkSession) logs(str ...string) {
	content := ""
	name := "log"

	if len(str) == 1 {
		content = str[0]
	} else if len(str) > 1 {
		content = str[0]
		name = str[1]
	}

	self.muGlobal.Lock()
	t := time.Now().Unix()
	logs(fm("%d > %12.12s %10.10s %9.9s > %s", t, self.Login, self.MyName, self.MyId, content), name)
	self.muGlobal.Unlock()
}

func (self *VkSession) logsErr(err error) {
	self.muGlobal.Lock()
	t := time.Now().Unix()
	logs(fm("%d > %12.12s %10.10s %9.9s > %v", t, self.Login, self.MyName, self.MyId, err), "error")
	self.muGlobal.Unlock()
}

type Methods struct {
	vk                  *VkSession
	hashSendMsg         map[int]string
	hashDelFriends      string
	postFromTargetGroup []string
	feedSession         string
	hashViewPost        string
	status              bool
	working             bool
	alreadyViewPosts    []string
	metaView            int
}

func InitMethods() *Methods {
	return &Methods{
		hashSendMsg:         make(map[int]string, 30),
		postFromTargetGroup: make([]string, 0, 300),
		working:             true,
		alreadyViewPosts:    make([]string, 0, 300),
		metaView:            255 + rand.Intn(925-255),
	}
}

func (self *Methods) checkStatus(r *Response) bool {
	if r.In("blockedHash") {
		self.status = false
		self.working = false

		self.vk.log("AKK BLOCKED!")
		self.vk.logs("AKK BLOCKED!")
		self.vk.wait.Done()
		return false

	} else {
		self.status = true
		return true
	}
}

func (self *Methods) LongPoll(act *Action) {
	var (
		err                             error
		ts, f                           int
		server, key, urlServer, longUrl string
		updates                         [][]byte
		res                             *Response
	)

	params := map[string]string{
		"act":  "a_get_key",
		"al":   "1",
		"gid":  "0",
		"im_v": "3",
		"uid":  self.vk.MyId}

	for self.working {
		self.vk.log("start long poll")
		res, err = self.vk.POST(im, params)
		if !res.Check("error long poll key", err) {
			res.Close()
			randSleep(15, 5)
			continue
		}

		j := res.NewJsArray("payload.[1]")
		res.Close()

		key = j.Get(0)
		urlServer = j.Get(1)
		server = j.Get(2)
		ts = j.Int(3)

		if key == "" || urlServer == "" || server == "" {
			randSleep(15, 5)
			continue
		}

		for self.working {
			longUrl = fm("%s/%s?act=a_check&key=%s&mode=202&ts=%d&version=9&wait=25", urlServer, server, key, ts)

			res, err = self.vk.GET(longUrl)
			if err != nil {
				res.Close()
				self.vk.logsErr(err)
				randSleep(15, 5)
				break
			}

			if f, _ = res.jsInt("failed"); f != 0 {
				randSleep(15, 5)
				res.Close()
				break
			}

			if updates, err = res.jsArr("updates"); err != nil {
				self.vk.logsErr(err)
				randSleep(15, 5)
				res.Close()
				break
			}

			//отправка в обработчик
			if len(updates) != 0 {
				for _, up := range updates {
					event := new(Event)
					event.pars(up, false, self.vk.re)
					if !event.empty {
						self.vk.logs(fm("%v", string(up)), "update")
						go act.eventProcessing(event)
					}
				}
			}
			if ts, err = res.jsInt("ts"); err != nil {
				self.vk.logsErr(err)
				randSleep(15, 5)
				res.Close()
				break
			}
			res.Close()
		}
	}
}

func (self *Methods) LongPollFeed(act *Action) {
	var (
		params                                        map[string]string
		res                                           *Response
		err                                           error
		e, r                                          [][]byte
		ts, ts1, ts2, key1, key2, serverUrl, finalKey string
	)

	for self.working {
		self.vk.log("start long poll feed")
		res, err = self.vk.GET("https://vk.com/id" + self.vk.MyId)
		if err != nil {
			self.vk.logsErr(err)
			res.Close()
			randSleep(15, 5)
			continue
		}

		key1 = res.FindFirst(",\"key\":\"", "\",\"uid\":")
		key2 = res.FindFirst("[{\"key\":\"", "\",\"ts\":")
		ts1 = res.FindFirst("\"timestamp\":", ",\"key\"")
		ts2 = res.FindFirst(",\"ts\":", "}]")
		serverUrl = res.FindFirst("\"server_url\":\"", "\",\"frame")
		res.Close()

		if key1 == "" || key2 == "" || ts1 == "" || ts2 == "" || serverUrl == "" {
			randSleep(15, 5)
			continue
		}

		serverUrl = reSubBackSlash.ReplaceAllString(serverUrl, "")
		finalKey = key1 + key2

		for self.working {
			ts = fm("%v_%v", ts1, ts2)

			params = map[string]string{
				"act":  "a_check",
				"id":   self.vk.MyId,
				"key":  finalKey,
				"ts":   ts,
				"wait": "25",
			}

			if res, err = self.vk.POST(serverUrl, params); err != nil {
				self.vk.log("error req Post", err)
				res.Close()
				randSleep(5, 5)
				break
			}
			if err != nil || res == nil {
				self.vk.logsErr(err)
				randSleep(15, 5)
				break
			}

			isFailed, _ := res.jsInt("failed")
			if isFailed != 0 {
				res.Close()
				randSleep(15, 5)
				break
			}

			//полученик слайса обновлений feed
			if r, err = res.jsArr(); err != nil || len(r) == 0 {
				self.vk.logsErr(err)
				res.Close()
				randSleep(15, 5)
				break
			}

			res.Close()

			//получение слайса всех event
			if e, err = jsArr(r[0], "events"); err != nil {
				self.vk.logsErr(err)
				randSleep(15, 5)
				break
			}

			//создание обработчиков
			if len(e) != 0 {
				for _, up := range e {
					up = delBackSlash(up)
					event := new(Event)
					event.pars(up, true, self.vk.re)
					if !event.empty {
						self.vk.logs(fm("%s", string(up)), "update_feed")
						go act.eventProcessing(event)
					}
				}
			}

			ts1, _ = jsStr(r[0], "ts")
			ts2, _ = jsStr(r[1], "ts")

			randSleep(1, 0)
		}
	}
}

func (self *Methods) setOnline() bool {
	urlStrings := []string{
		fm("https://vk.com/id%v", self.vk.MyId),
		"https://vk.com/feed",
		"https://vk.com/im",
		"https://vk.com/groups",
		fm("https://vk.com/audios%v", self.vk.MyId),
		"https://vk.com/video",
		fm("https://vk.com/id%v", strconv.Itoa(rand.Intn(591626759)+100000)),
	}
	if isRand(50) {
		res, err := self.vk.GET(urlStrings[rand.Intn(len(urlStrings))])
		if err != nil {
			res.Close()
			self.vk.logsErr(err)
			return false
		}
		self.checkStatus(res)
		res.Close()
		randSleep(10, 1)
	}

	params := map[string]string{
		"act":  "a_onlines",
		"al":   "1",
		"peer": "",
	}
	return !self.simpleMethod(im, "error set online", params)
}

func (self *Methods) simpleMethod(url, msg string, params map[string]string) bool {
	res, err := self.vk.POST(url, params)
	defer res.Close()
	return res.Check(msg, err)
}

func (self *Methods) editMsg(userId int, msg string, atta [][]string, msgId int, hashMsg string) bool {
	media := ""
	if len(atta) != 0 {
		media = fmt.Sprintf("%v", atta)
	}
	params := map[string]string{
		"act":    "a_edit_message",
		"al":     "1",
		"gid":    "0",
		"hash":   hashMsg,
		"id":     st(msgId),
		"im_v":   "3",
		"media":  media,
		"msg":    msg,
		"peerId": st(userId),
	}
	return self.simpleMethod(im, "error edit message", params)
}

func (self *Methods) readMsg(userId, msgId int, hashMsg string) bool {
	params := map[string]string{
		"act":    "a_mark_read",
		"al":     "1",
		"gid":    "0",
		"hash":   hashMsg,
		"ids[0]": st(msgId),
		"im_v":   "3",
		"peer":   st(userId)}
	return self.simpleMethod(im, "error read message", params)
}

func (self *Methods) setTyping(userId int, hashMsg string) bool {
	params := map[string]string{
		"act":  "a_activity",
		"al":   "1",
		"gid":  "0",
		"hash": hashMsg,
		"im_v": "3",
		"peer": st(userId),
		"type": "typing"}
	return self.simpleMethod(im, "error typing", params)
}

func (self *Methods) getHashSend(userId int) (string, error) {
	self.vk.muLocal.Lock()
	hashMsg, ok := self.hashSendMsg[userId]
	self.vk.muLocal.Unlock()
	if !ok {
		params := map[string]string{
			"act":      "a_start",
			"al":       "1",
			"block":    "true",
			"gid":      "0",
			"history":  "true",
			"im_v":     "3",
			"msgid":    "false",
			"peer":     st(userId),
			"prevpeer": "0"}
		res, err := self.vk.POST(im, params)
		defer res.Close()
		if !res.Check("error get IM", err) {
			res.Close()
			return "", err
		}
		h := res.FindFirst("\"hash\":\"", "\",")
		if h != "" {
			hashMsg = h
			self.vk.muLocal.Lock()
			self.hashSendMsg[userId] = hashMsg
			self.vk.muLocal.Unlock()
		}
	}
	return hashMsg, nil
}

func (self *Methods) setSend(userId int, msg string, atta [][]string, hashMsg string) (int, error) {
	params := map[string]string{
		"act":        "a_send",
		"al":         "1",
		"entrypoint": "",
		"gid":        "0",
		"guid":       st(int(time.Now().UnixNano()))[0:15],
		"hash":       hashMsg,
		"im_v":       "3",
		"media":      "",
		"msg":        msg,
		"random_id":  st(rand.Intn(100000000) + 10000000),
		"to":         st(userId)}

	if atta != nil {
		temp := make([]string, 0, 10)
		for _, val := range atta {
			temp = append(temp, fmt.Sprintf("%s:%s:undefined", val[0], val[1]))
		}
		params["media"] = strings.Join(temp, ",")
	}
	res, err := self.vk.POST(im, params)
	defer res.Close()
	if !res.Check("Error send message", err) {
		return 0, err
	}
	i, err := res.jsInt("payload.[1].[0].msg_id")
	if err != nil {
		return -1, err
	}
	return i, nil
}

func (self *Methods) SendMessage(userId int, msg string, atta [][]string,
	msgId int, readMsg, typing bool) (int, error) {
	var hashMsg string
	var err error

	randSleep(15, 5)

	hashMsg, err = self.getHashSend(userId)
	if err != nil || len(hashMsg) == 0 {
		return 0, err
	}

	randSleep(50, 5)

	if readMsg {
		self.readMsg(userId, msgId, hashMsg)
	}

	if len(msg) == 0 {
		self.vk.log("Not msg")
		return 0, nil
	}

	randSleep(10, 1)

	if typing {
		lenMsg := len([]rune(msg))/15 + 1
		for x := 0; x < lenMsg; x++ {
			self.setTyping(userId, hashMsg)
			randSleep(5, 3)
		}
	}
	randSleep(5, 3)

	msgId = 0
	msgId, err = self.setSend(userId, msg, atta, hashMsg)
	if err != nil {
		return 0, err
	}

	if isRand(15) {
		randSleep(30, 1)
		self.editMsg(userId, msg, atta, msgId, hashMsg)
	}

	return msgId, nil
}

func (self *Methods) Subscribe(ownerId int, hashS string) (bool, error) {
	var hashSL []string
	var params map[string]string
	var group bool
	var urlP string

	randSleep(5, 5)

	res, err := self.vk.GET(fm("https://vk.com/club%d", ownerId))
	defer res.Close()
	if err != nil {
		self.vk.logsErr(err)
		return false, err
	}

	if len(hashS) == 0 {
		hashSL = res.Find("\"enterHash\":\"", "\",")
		if len(hashSL) != 0 {
			hashS = hashSL[0]
		}
	}

	group = false
	if len(hashS) == 0 {
		hashSL = res.Find("Groups.enter(this, ", "',")
		if len(hashSL) == 0 {
			self.vk.log("probably already signed 1")
			return false, nil
		}
		hashSL = reFindSubHash.FindAllString(hashSL[0], -1)
		if len(hashSL) != 2 {
			self.vk.log("probably already signed 2")
			return false, nil
		}
		group = true
		hashS = hashSL[1]
	}

	randSleep(20, 10)

	if group {
		params = map[string]string{
			"act":  "enter",
			"al":   "1",
			"gid":  st(ownerId),
			"hash": hashS,
		}
		urlP = al_groups
	} else {
		params = map[string]string{
			"act":  "a_enter",
			"al":   "1",
			"pid":  st(ownerId),
			"hash": hashS,
		}
		urlP = al_public
	}

	if self.simpleMethod(urlP, "subscribe", params) {
		self.vk.logs(fmt.Sprintf("Вступил в сообщество: %d", ownerId))
		return true, nil
	} else {
		self.vk.logs(fmt.Sprintf("Вступить в сообщество не получилось: %d", ownerId))
		return false, nil
	}
}

func (self *Methods) Leave(ownerId int, hashL string) (bool, error) {
	var params map[string]string
	var hashLeave []string
	var urlP string

	randSleep(5, 3)

	if len(hashL) == 0 {
		res, err := self.vk.GET(fmt.Sprintf("https://vk.com/club%d", ownerId))
		defer res.Close()
		if err != nil {
			self.vk.logsErr(err)
			return false, err
		}

		hashLeave = res.Find("\"enterHash\":\"", "\",")

		if len(hashLeave) == 0 {
			hashLeave = res.Find("Groups.leave('page_actions_btn',", "')")
			if len(hashLeave) == 0 {
				self.vk.log("probably already leave the community", ownerId)
				return false, nil
			}

			hashLeave = reFindSubHash.FindAllString(hashLeave[0], -1)
			if len(hashLeave) != 2 {
				self.vk.log("probably already leave the community", ownerId)
				return false, nil
			}

			params = map[string]string{
				"act":  "leave",
				"al":   "1",
				"gid":  st(ownerId),
				"hash": hashLeave[1],
			}
			urlP = al_groups
		} else {
			params = map[string]string{
				"act":  "a_leave",
				"al":   "1",
				"pid":  st(ownerId),
				"hash": hashLeave[0],
			}
			urlP = al_public
		}
	} else {
		params = map[string]string{
			"act":  "list_leave",
			"al":   "1",
			"gid":  st(ownerId),
			"hash": hashL,
		}
		urlP = al_groups
	}

	if self.simpleMethod(urlP, "error leave group", params) {
		self.vk.logs(fmt.Sprintf("Покинул сообщество: %d", ownerId))
		return true, nil
	} else {
		self.vk.logs(fmt.Sprintf("Не удалось покинуть сообщество: %d", ownerId))
		return false, nil
	}
}

func (self *Methods) getHashPost(idPost string) map[string][]string {
	res, err := self.vk.GET(fm("https://vk.com/%s", idPost))
	defer res.Close()
	if err != nil {
		self.vk.logsErr(err)
		return nil
	}

	hashLikeList := res.Find("Likes.toggle(this, event, ", "');")
	if len(hashLikeList) != 0 {
		hashLikeList = reFindSubHash.FindAllString(hashLikeList[0], -1)
	} else {
		hashLikeList = []string{}
	}

	hashCommentList := res.Find("post_hash\":\"", "\",\"")
	commentIdList := res.Find("<div id=\"wpt", "\">")

	hashSpamList := res.Find("wall.markAsSpam(this,", ");")
	if len(hashSpamList) != 0 {
		hashSpamList = reFindSubHash.FindAllString(hashSpamList[0], -1)
	} else {
		hashSpamList = []string{}
	}

	out := map[string][]string{
		"like":        hashLikeList,
		"comment":     hashCommentList,
		"comment_ids": commentIdList,
		"spam":        hashSpamList,
	}
	return out
}

func (self *Methods) getInfoView(idPost string) int {
	params := map[string]string{
		"act":    "a_get_stats",
		"al":     "1",
		"object": idPost,
		"views":  "1",
	}
	res, err := self.vk.POST(like, params)
	defer res.Close()
	if !res.Check("error get views", err) {
		return -1
	}

	j, err := res.jsStr("payload.[1].[1]")
	if err != nil {
		return -1
	}

	sList := reFindD.FindAllString(j, 1)
	if len(sList) == 0 {
		return -1
	}

	i, err := strconv.Atoi(sList[0])
	if err != nil {
		return -1
	}

	randSleep(5, 1)

	return i
}

func (self *Methods) parsInfo(res *Response) map[string]int {
	l := res.Find("Likes.update(", ");")
	if len(l) == 0 {
		return nil
	}
	l = reFigureScope.FindAllString(l[0], -1)
	if len(l) == 0 {
		return nil
	}

	var v map[string]int
	p := reSubBackSlash.ReplaceAllString(l[0], "")
	err := json.Unmarshal([]byte(p), &v)
	if err != nil {
		self.vk.logsErr(err)
		return nil
	}

	return v
}

func (self *Methods) getInfoLike(idPost string) (int, bool) {
	params := map[string]string{
		"act":       "a_get_stats",
		"al":        "1",
		"has_share": "1",
		"object":    idPost,
	}
	res, err := self.vk.POST(like, params)
	defer res.Close()
	if !res.Check("error get info like post", err) {
		return -1, true
	}

	v := self.parsInfo(res)
	if v == nil {
		return -1, true
	}

	randSleep(5, 1)

	return v["like_num"], v["like_my"] == 1
}

func (self *Methods) getInfoRepost(idPost string) (int, bool) {
	params := map[string]string{
		"act":       "a_get_stats",
		"al":        "1",
		"has_share": "1",
		"object":    idPost,
		"published": "1",
	}
	res, err := self.vk.POST(like, params)
	defer res.Close()
	if !res.Check("error get info repost", err) {
		return -1, true
	}

	v := self.parsInfo(res)
	if v == nil {
		return -1, true
	}

	randSleep(5, 1)

	return v["share_num"], v["share_my"] == 1
}

func (self *Methods) getHashRepost(idPost string) string {
	params := map[string]string{
		"act":    "publish_box",
		"al":     "1",
		"object": idPost}
	res, err := self.vk.POST(like, params)
	defer res.Close()
	if !res.Check("error repost publish_box", err) {
		return ""
	}

	hashR := res.FindFirst("shHash: '", "',")

	randSleep(5, 1)

	params = map[string]string{
		"act":  "a_json_friends",
		"al":   "1",
		"from": "imwrite",
		"str":  "",
	}

	resp, _ := self.vk.POST(hints, params)
	resp.Close()

	randSleep(15, 3)

	return hashR
}

func (self *Methods) setRepost(idPost, hashR, msg string) bool {
	params := map[string]string{
		"Message":            msg,
		"act":                "a_do_publish",
		"al":                 "1",
		"close_comments":     "0",
		"friends_only":       "0",
		"from":               "box",
		"hash":               hashR,
		"list":               "",
		"mark_as_ads":        "0",
		"mute_notifications": "0",
		"object":             idPost,
		"ret_data":           "1",
		"to":                 "0"}
	return self.simpleMethod(like, "error repost", params)
}

func (self *Methods) setLike(idPost, hashL, likeFrom string) bool {
	params := map[string]string{
		"act":    "a_do_like",
		"al":     "1",
		"from":   likeFrom, // 'wall_one', wall_page, feed_recent
		"hash":   hashL,
		"object": idPost,
		"wall":   "2",
	}
	return self.simpleMethod(like, "error set like", params)
}

func (self *Methods) Like(idPost, hashL, likeFrom string) bool {
	self.vk.log("start like", idPost)

	if len(hashL) == 0 {
		likeFrom = "wall_one"
		h, ok := self.getHashPost(idPost)["like"]
		if ok && len(h) >= 2 {
			hashL = h[1]
		} else {
			self.vk.logs(fmt.Sprintf("Ошибка лайка: %s", idPost))
			return false
		}
		randSleep(10, 5)
	}

	if isRand(40) {
		self.getInfoView(idPost)
	}
	if isRand(40) {
		self.getInfoRepost(idPost)
	}
	_, my := self.getInfoLike(idPost)

	s := false
	m := ""
	if !my {
		s = self.setLike(idPost, hashL, likeFrom)
	}

	if s {
		m = "Лайк поставлен:"
	} else {
		m = "Лайк не поставлен:"
	}

	self.vk.logs(fm("%s %s", m, idPost))
	return s
}

func (self *Methods) getFeedPost(res *Response, count int) []byte {
	var resp *Response
	var err error
	var from, section, subsection, offset string
	var params map[string]string

	feedBody := make([]byte, 0, count*65000)

	for _, val := range res.getByte(-1) {
		feedBody = append(feedBody, val)
	}

	b := []byte("{" + res.FindFirst("feed.init({", ",\"all_shown_text\"") + "}")

	from, _ = js(b, "from")
	section, _ = js(b, "section")
	subsection, _ = js(b, "subsection")

	params = map[string]string{
		"al":         "1",
		"al_ad":      "0",
		"from":       from,
		"more":       "1",
		"offset":     "10",
		"part":       "1",
		"section":    section,
		"subsection": subsection,
	}

	for num := 0; num <= 100; num += 10 {
		if num > count-10 {
			break
		}

		resp, err = self.vk.POST("https://vk.com/al_feed.php?sm_news=", params)
		if !resp.Check("error get more from feed", err) {
			self.vk.log(err)
			break
		}

		b, _ = resp.jsAsByte("payload.[1].[0]")

		from, _ = js(b, "from")
		section, _ = js(b, "section")
		subsection, _ = js(b, "subsection")
		offset, _ = js(b, "offset")

		body, _ := resp.jsAsByte("payload.[1].[1]")
		resp.Close()

		for i := range body {
			feedBody = append(feedBody, body[i])
		}

		params = map[string]string{
			"al":         "1",
			"al_ad":      "0",
			"from":       from,
			"more":       "1",
			"offset":     offset,
			"part":       "1",
			"section":    section,
			"subsection": subsection,
		}

		randSleep(10, 10)
	}

	return feedBody
}

func (self *Methods) getPostFrom(ownerId string, res *Response, count int) ([][]string, error) {
	var err error
	var hashLikeFeedFinal = make([][]string, 0, 20)
	var response = make([][]string, 0, 20)
	var adsGroup = make([]string, 0, 10)
	var fixPost string

	if ownerId == "group" && res != nil {
		self.feedSession = "na"
	} else if ownerId == "feed" && res != nil {
		self.feedSession = res.FindFirst("feed_session_id\":", ",\"feedba")
	} else if ownerId != "" && res == nil {
		res, err = self.vk.GET("https://vk.com/" + ownerId)
		defer res.Close()
		if err != nil {
			return nil, err
		}
		self.feedSession = "na"
	} else if ownerId == "" && res == nil {
		res, err = self.vk.GET("https://vk.com/feed")
		if err != nil {
			return nil, err
		}
		self.feedSession = res.FindFirst("feed_session_id\":", ",\"feedba")

		if count > 10 {
			temp := self.getFeedPost(res, count)
			res.Close()
			res = self.vk.response.Write(temp)
			defer res.Close()
		} else {
			defer res.Close()
		}
	} else {
		return nil, nil
	}

	self.hashViewPost = res.FindFirst("post_view_hash=\"", "\".")

	hashLikeFeed := res.Find("Likes.toggle(this, event, '", ");")
	for _, val := range hashLikeFeed {
		tempStringList := reFindSubHash.FindAllString(val, -1)
		if len(tempStringList) != 2 {
			continue
		}
		hashLikeFeedFinal = append(hashLikeFeedFinal, tempStringList)
	}

	adsPost := res.Find("promoted_post post_link\" href=\"/", "\" onclick")

	for _, val := range res.Find("class=\"post_content\"", "class=\"replies\"") {
		x := rePostId.FindAllString(val, -1)
		if len(x) != 0 {
			adsGroup = append(adsGroup, x[0])
		}
	}

	fixPostList := res.Find("id=\"post", "\" class=\"_post post all own post_fixed")

	if len(fixPostList) != 0 {
		fixPost = "wall" + fixPostList[0]
	}

	for _, val := range hashLikeFeedFinal {
		if strings.Contains(val[0], "wall-") && !IsIn(val[0], adsPost) && !IsIn(val[0], adsGroup) && val[0] != fixPost {
			response = append(response, []string{val[0], val[1]})
		}
	}
	return response, nil
}

func (self *Methods) Repost(idPost, msg string) bool {
	self.vk.log("repost", idPost, msg)

	randSleep(20, 15)

	_, my := self.getInfoRepost(idPost)
	if my {
		self.vk.logs("уже делал репост этого поста: " + idPost)
		return false
	}
	if isRand(40) {
		self.getInfoView(idPost)
	}
	if isRand(40) {
		self.getInfoLike(idPost)
	}

	hashR := self.getHashRepost(idPost)
	if len(hashR) == 0 {
		self.vk.logs("Ошибка репоста: " + idPost)
		return false
	}
	s := self.setRepost(idPost, hashR, msg)

	m := ""
	if s {
		m = "Репост выполнен:"
	} else {
		m = "Репост не выполнен:"
	}

	self.vk.logs(fm("%s %s", m, idPost))

	return s
}

func (self *Methods) friendDecline(userId int, actionHash string) bool {
	params := map[string]string{
		"act":          "remove",
		"al":           "1",
		"from_section": "requests",
		"hash":         actionHash,
		"mid":          st(userId),
		"report_spam":  "1",
	}
	if self.simpleMethod(al_friends, "error friend_decline", params) {
		self.vk.logs(fmt.Sprintf("Успешная отмена заявки в друзья: %d", userId))
		return true
	} else {
		self.vk.logs(fmt.Sprintf("Ошибка отмены заявки в друзья: %d", userId))
		return false
	}
}

func (self *Methods) friendAccept(userId int, actionHash string) bool {
	params := map[string]string{
		"act":         "add",
		"al":          "1",
		"hash":        actionHash,
		"mid":         st(userId),
		"request":     "1",
		"select_list": "1",
	}
	if self.simpleMethod(al_friends, "error friend_decline", params) {
		self.vk.logs(fmt.Sprintf("Успешно принял заявку в друзья: %d", userId))
		return true
	} else {
		self.vk.logs(fmt.Sprintf("Ошибка принятия заявки в друзья: %d", userId))
		return false
	}
}

func (self *Methods) delComment() {
	//TODO
}

func (self *Methods) getImageComment() {
	//TODO
}

func (self *Methods) postWall() {
	//TODO
}

func (self *Methods) parsSearchGroups() {
	//TODO
}

func (self *Methods) searchGroups() {
	//TODO
}

func (self *Methods) spam() {
	//TODO
}

func (self *Methods) commentPost(idPost, msg string, replyToUser int,
	replyToMsg, hashComment string) bool {
	self.vk.log("comment post", idPost, msg)

	if msg == "" {
		self.vk.log("not message text", idPost)
		return false
	}

	if hashComment == "" {
		hash, ok := self.getHashPost(idPost)["comment"]
		if !ok || len(hash) == 0 {
			self.vk.log("not hash comment", idPost)
			return false
		}
		hashComment = hash[0]
		randSleep(15, 5)
	}

	idPostList := reFindD.FindAllString(idPost, -1)
	if len(idPostList) != 2 {
		self.vk.log("id post error", idPost)
		return false
	}

	params := map[string]string{
		"Message": msg,
		//"_ads_group_id": idPostList[0],
		"act":           "post",
		"al":            "1",
		"from":          "",
		"from_oid":      self.vk.MyId,
		"hash":          hashComment,
		"need_last":     "0",
		"only_new":      "1",
		"order":         "asc",
		"ref":           "wall_page",
		"reply_to":      fmt.Sprintf("%s_%s", idPostList[0], idPostList[1]),
		"reply_to_msg":  replyToMsg,
		"reply_to_user": st(replyToUser),
		"type":          "own",
	}

	if self.simpleMethod(al_wall, "error set comment", params) {
		self.vk.logs(fmt.Sprintf("Новый комментарий: %s, %s", idPost, msg))
		return true
	} else {
		self.vk.logs(fmt.Sprintf("Ошибка комментария к: %s, %s", idPost, msg))
		return false
	}
}

func (self *Methods) commentPhoto(idPhoto, msg, hashC string, fromId int) bool {
	self.vk.log("comment photo", idPhoto)

	replyToMsg := ""

	if hashC == "" {
		res, err := self.vk.GET("https://vk.com/%s" + idPhoto)
		defer res.Close()
		if !res.Check("Error get"+idPhoto+"for comment photo", err) {
			return false
		}

		hashC = res.FindFirst(idPhoto+"', '", "'")
		randSleep(15, 5)

		s := strings.Split(idPhoto, "_")
		if len(s) == 2 {
			p := self.vk.MyId + "_photo" + strings.Split(idPhoto, "_")[1]
			r := res.FindFirst(p, st(fromId))
			replyToMsg = reD.ReplaceAllString(r, "")
		}
	}

	params := map[string]string{
		"Message":    msg,
		"act":        "post_comment",
		"al":         "1",
		"from_group": "",
		"hash":       hashC,
		"photo":      reSubS.ReplaceAllString(idPhoto, ""),
		"reply_to":   replyToMsg,
	}

	if len(hashC) != 0 && self.simpleMethod(al_photos, "error comment photo", params) {
		self.vk.logs(fmt.Sprintf("Новый комментарий к фото: %s, %s", idPhoto, msg))
		return true
	} else {
		self.vk.logs(fmt.Sprintf("Ошибка комментария к фото: %s, %s", idPhoto, msg))
		return false
	}
}

func (self *Methods) viewPost(idPosts []string, target []string, randomView int) bool {
	var finIdPosts = make([]string, 0, 20)
	var data, meta, pref string

	d := []string{
		"-1", "2298", "-1", "1384", "-1", "3065", "-1",
	}
	for index := range idPosts {
		if (isRand(randomView) || IsIn(idPosts[index], target)) && IsIn(idPosts[index], self.alreadyViewPosts) {
			finIdPosts = append(finIdPosts, idPosts[index])
			self.alreadyViewPosts = append(self.alreadyViewPosts, idPosts[index])
		}
	}
	if len(self.alreadyViewPosts) > 301 {
		temp := make([]string, 0, 300)
		l := len(self.alreadyViewPosts)
		for _, val := range self.alreadyViewPosts[l/2 : l-1] {
			temp = append(temp, val)
		}
		self.alreadyViewPosts = temp
	}
	if self.feedSession == "na" {
		pref = "_c"
	} else {
		pref = []string{"_rf", "_tf"}[rand.Intn(2)]
	}
	for index, val := range finIdPosts {
		a := reFindD.FindAllString(val, -1)

		data += fmt.Sprintf("%s%s%s:%s:%d:%s;", a[0], pref, a[1], d[rand.Intn(len(d))], index, self.feedSession)
		meta += fmt.Sprintf("%s:%d:%d;", val, rand.Intn(600)+300, self.metaView)
	}
	params := map[string]string{
		"act":  "seen",
		"al":   "1",
		"data": data,
		"hash": self.hashViewPost,
		"meta": meta}
	return self.simpleMethod(al_page, "error view posts", params)
}

func (self *Methods) getNewFriendList() (map[string][]string, error) {
	var res *Response
	var err error
	var check, temp []string
	var hashA = make([][]string, 0, 15)
	var hashD = make([][]string, 0, 15)
	var responseMap = make(map[string][]string, 15)

	res, err = self.vk.GET("https://vk.com/friends?section=requests")
	defer res.Close()
	if err != nil {
		return nil, err
	}

	check = res.Find("section=requests\" class=\"", "\"onclick=\"return Friends")
	if len(check) >= 2 && check[1] != "ui_tab ui_tab_sel" {
		return nil, nil
	}

	for _, val := range res.Find("\" onclick=\"Friends.acceptRequest(", "', this)\">") {
		temp = reFindSubHash.FindAllString(val, -1)
		if len(temp) != 2 {
			continue
		}
		hashA = append(hashA, temp)
	}

	for _, val := range res.Find("\" onclick=\"Friends.declineRequest(", "', this)\">") {
		temp = reFindSubHash.FindAllString(val, -1)
		if len(temp) != 2 {
			continue
		}
		hashD = append(hashD, temp)
	}

	if len(hashD) != len(hashA) {
		self.vk.log("get new friend list error != len")
		return nil, nil
	}

	for index := range hashA {
		responseMap[hashA[index][0]] = []string{hashA[index][1], hashD[index][1]}
	}

	if len(hashA) == 15 {
		responseMap = self.getFriendsRequests(responseMap, 0)
	}

	return responseMap, nil
}

func (self *Methods) getFriendsRequests(value map[string][]string, offset int) map[string][]string {
	var res *Response
	var err error
	var requests [][]byte

	for x := 0; x < 5; x++ {
		params := map[string]string{
			"act":     "get_section_friends",
			"al":      "1",
			"gid":     "0",
			"id":      self.vk.MyId,
			"offset":  st(offset),
			"section": "requests",
		}
		res, err = self.vk.POST(friends, params)
		if !res.Check("error get_section_friends", err) {
			res.Close()
			return value
		}

		j, _ := res.jsAsByte("payload.[1].[0]")
		requests, err = jsArr(delBackSlash(j), "requests")
		if len(requests) == 0 || err != nil {
			self.vk.logsErr(err)
			res.Close()
			return value
		}

		for i := range requests {
			id, _ := js(requests[i], "[0]")
			hashA, _ := js(requests[i], "[9].[0]")
			hashD, _ := js(requests[i], "[9].[1]")

			if id != "" || hashA != "" || hashD != "" {
				value[id] = []string{hashA, hashD}
			}
		}
		if len(requests) == 15 {
			offset += 15
		} else {
			break
		}
		res.Close()
		randSleep(10, 2)
	}
	return value
}

func (self *Methods) delFriend(userId int, toBlackList bool) (bool, error) {
	var res *Response
	var err error

	if self.hashDelFriends == "" {
		res, err = self.vk.GET("https://vk.com/friends")
		defer res.Close()
		if err != nil || res == nil {
			return false, err
		}
		self.hashDelFriends = res.FindFirst("\"userHash\":\"", "\",\"")
		if self.hashDelFriends == "" {
			return false, err
		}
	}

	randSleep(20, 10)

	params := map[string]string{
		"act":  "remove",
		"al":   "1",
		"hash": self.hashDelFriends,
		"mid":  st(userId),
	}

	s := self.simpleMethod(al_friends, "error del friend", params)
	if s {
		self.vk.logs(fm("Удалил друга: %d", userId))
	} else {
		self.vk.logs(fm("Не смог удалил друга: %d", userId))
	}

	if toBlackList && s {
		_, _ = self.addUserToBlackList(userId)
	}

	return s, nil
}

func (self *Methods) addUserToBlackList(userId int) (bool, error) {
	randSleep(15, 5)

	res, err := self.vk.GET(fm("https://vk.com/id%d", userId))
	if err != nil || res == nil {
		res.Close()
		self.vk.logsErr(err)
		return false, err
	}

	isFriend := res.In("Profile.frDropdownPreload.pbind")
	isSubscriber := res.In("profile_am_subscribed")

	if isFriend && !isSubscriber {
		_, _ = self.delFriend(userId, false)

		res.Close()
		randSleep(30, 10)

		res, err = self.vk.GET(fm("https://vk.com/id%d", userId))
		if err != nil {
			res.Close()
			return false, err
		}
	}

	hashB := res.FindFirst("Profile.toggleBlacklist(this, '", "', event")
	res.Close()

	if hashB == "" {
		self.vk.log("error not hash black list")
		return false, nil
	}

	randSleep(10, 5)

	params := map[string]string{
		"act":  "a_add_to_bl",
		"al":   "1",
		"from": "profile",
		"hash": hashB,
		"id":   st(userId)}

	s := self.simpleMethod(al_settings, "error_add_user_to_black_list", params)
	if s {
		self.vk.logs(fm("Добавил в ЧС: %d", userId))
	} else {
		self.vk.logs(fm("Неудалось добавить в ЧС: %d", userId))
	}
	return s, nil
}

func (self *Methods) delOutRequest(userId int, hashD string) (bool, error) {
	params := map[string]string{
		"act":          "remove",
		"al":           "1",
		"from_section": "out_requests",
		"hash":         hashD,
		"mid":          st(userId),
		"report_spam":  "1",
	}
	s := self.simpleMethod(al_friends, "error del out requests", params)
	if s {
		self.vk.logs(fm("Удалил исходящюю заявку в друзья к: %d", userId))
	} else {
		self.vk.logs(fm("Не удалось удалить исходящюю заявку в друзья к: %d", userId))
	}
	return s, nil
}

func (self *Methods) getSectionFriendsOutRequests(offset int) ([][]string, error) {
	var responseList = make([][]string, 0, 15)
	var requests [][]byte
	var j []byte
	var err error
	var res *Response
	var id, hashD string

	params := map[string]string{
		"act":     "get_section_friends",
		"al":      "1",
		"gid":     "0",
		"id":      self.vk.MyId,
		"offset":  st(offset),
		"section": "out_requests",
	}
	res, err = self.vk.POST(al_friends, params)
	defer res.Close()
	if !res.Check("error get SectionFriendsOutRequests", err) {
		return responseList, err
	}

	j, _ = res.jsAsByte("payload.[1].[0]")
	requests, err = jsArr(delBackSlash(j), "out_requests")
	if len(requests) == 0 || err != nil {
		return responseList, err
	}

	for i := range requests {
		id, _ = js(requests[i], "[0]")
		hashD, _ = js(requests[i], "[10].[1]")
		if id == "" || hashD == "" {
			continue
		}
		responseList = append(responseList, []string{id, hashD})
	}
	return responseList, nil
}

func (self *Methods) DelOutRequests(toBlackList bool) error {
	var hashO = make([][]string, 0, 15)
	var temp = make([]string, 0, 2)
	var offset = 0
	var count = 0
	var ok bool

	randSleep(20, 10)
	self.vk.logs("Старт удаления исходящих заявок в друзья")

	data, err := self.vk.GET("https://vk.com/friends?section=out_requests")
	defer data.Close()
	if err != nil || data == nil {
		return err
	}

	for _, val := range data.Find("\" onclick=\"Friends.declineRequest(", "', this)\">") {
		temp = reFindSubHash.FindAllString(val, -1)
		if len(temp) != 2 {
			continue
		}
		hashO = append(hashO, temp)
	}

	if len(hashO) == 0 || !data.In("flat_button button_small fl_r") {
		self.vk.logs("Нет исходящих заявок в друзья")
		return nil
	}

	for x := 0; x < 5; x++ {
		if len(hashO) == 0 {
			break
		}

		offset += len(hashO)

		for _, user := range hashO {
			id, _ := strconv.Atoi(user[0])
			if id == 0 {
				continue
			}
			ok, _ = self.delOutRequest(id, user[1])
			if ok {
				count++
			}
			randSleep(25, 10)
			if toBlackList {
				_, _ = self.addUserToBlackList(id)
			}
		}

		randSleep(20, 10)

		hashO, err = self.getSectionFriendsOutRequests(offset)
		if err != nil {
			self.vk.logsErr(err)
			break
		}
		randSleep(10, 3)
	}
	self.vk.logs(fm("Удалил исходящих заявок в друзья: %d", count))
	return nil
}

func (self *Methods) getHashDialogs() (map[string][]string, error) {
	randSleep(3, 1)
	outMap := make(map[string][]string, 20)

	res, err := self.vk.GET("https://vk.com/im")
	defer res.Close()
	if err != nil || res == nil {
		return nil, err
	}

	imInit := res.FindFirst("IM.init({", "})")
	if imInit == "" {
		return nil, nil
	}
	imInit = fm("{%s}", imInit)

	ar, err := jsObj([]byte(imInit), "tabs")
	if err != nil || len(ar) == 0 {
		return nil, err
	}

	for key, val := range ar {
		hashD, _ := js(val, "hash")
		href, _ := js(val, "href")

		h := strings.Split(href, "=c")
		href = h[len(h)-1]

		outMap[key] = []string{hashD, href}
	}
	return outMap, nil
}

func (self *Methods) delDialog(peerId, hashD string) (bool, error) {
	if hashD == "" {
		mapHashD, err := self.getHashDialogs()
		if err != nil {
			return false, err
		}

		temp, ok := mapHashD[peerId]
		if !ok {
			self.vk.logs(fm("Ошибка удаления диалога: %s", peerId))
			return false, nil
		}
		hashD = temp[0]
	}

	randSleep(10, 3)

	params := map[string]string{
		"act":  "a_delete_dialog",
		"al":   "1",
		"gid":  "0",
		"hash": hashD,
		"im_v": "3",
		"peer": peerId}

	s := self.simpleMethod(im, "error del dialog", params)
	if s {
		self.vk.logs(fm("Удалил диалог: %s", peerId))
	} else {
		self.vk.logs(fm("Ошибка удаления диалога: %s", peerId))
	}
	return s, nil
}

func (self *Methods) leaveChat(peerId, hashL, chatId string) (bool, error) {
	if hashL == "" || chatId == "" {
		mapHashD, err := self.getHashDialogs()
		if err != nil {
			return false, err
		}
		temp, ok := mapHashD[peerId]
		if !ok {
			self.vk.logs(fm("Ошибка выхода из диалога: %s %s", peerId, chatId))
			return false, nil
		}
		hashL = temp[0]
		chatId = temp[1]
	}

	randSleep(10, 3)

	params := map[string]string{
		"_smt": "im:22",
		"act":  "a_leave_chat",
		"al":   "1",
		"chat": chatId,
		"gid":  "0",
		"hash": hashL,
		"im_v": "3",
	}

	s := self.simpleMethod(im, "error leave chat", params)
	if s {
		self.vk.logs("Покинул чат: %s, %s", peerId, chatId)
	} else {
		self.vk.logs("Неудалось покинуть чат: %s, %s", peerId, chatId)
	}
	return s, nil
}

func (self *Methods) delHistoryDialog(peerId, hashH string) (bool, error) {
	if hashH == "" {
		mapHashDialog, err := self.getHashDialogs()
		if err != nil {
			return false, err
		}
		temp, ok := mapHashDialog[peerId]
		if !ok {
			self.vk.logs(fm("Ошибка удаления истории диалога: %s", peerId))
			return false, nil
		}
		hashH = temp[0]
	}

	randSleep(10, 3)

	params := map[string]string{
		"act":  "a_flush_history",
		"al":   "1",
		"from": "im",
		"gid":  "0",
		"hash": hashH,
		"id":   peerId,
		"im_v": "3"}

	s := self.simpleMethod(im, "error del history", params)
	if s {
		self.vk.logs(fm("Удалил историю диалога: %s", peerId))
	} else {
		self.vk.logs(fm("Не смог удалить историю диалога: %s", peerId))
	}
	return s, nil
}

func (self *Methods) getGroups(userId string) (*[]Group, error) {
	var listGroup = make([]Group, 0, 100)
	var group Group

	params := map[string]string{
		"act": "get_list",
		"al":  "1",
		"mid": userId,
		"tab": "groups"}
	res, err := self.vk.POST(al_groups, params)
	defer res.Close()
	if !res.Check("error get list group", err) {
		return nil, err
	}

	ar, _ := res.jsArr("payload.[1].[0]")

	for i := range ar {
		j := NewJsArray(ar[i])
		group = Group{
			Id:        j.Get(2),
			Name:      j.Get(0),
			UrlGroup:  j.Get(3),
			UrlPhoto:  j.Get(4),
			HashLeave: j.Get(7),
		}
		listGroup = append(listGroup, group)
	}
	return &listGroup, nil
}

func (self *Methods) getFriends(userId string) (*[]User, error) {
	var listFrieds = make([]User, 0, 300)

	if userId != self.vk.MyId {
		_, _ = self.vk.GET("https://vk.com/id" + userId)
		randSleep(45, 15)
	}

	params := map[string]string{
		"act": "load_friends_silent",
		"al":  "1",
		"gid": "0",
		"id":  userId,
	}
	res, err := self.vk.POST(al_friends, params)
	defer res.Close()
	if !res.Check("error get list friends", err) {
		return nil, err
	}

	ar, _ := res.jsArr("payload.[1].[0].all")

	for i := range ar {
		j := NewJsArray(ar[i])

		temp := User{
			Id:       j.Get(0),
			Name:     j.Get(5),
			UrlPhoto: j.Get(1),
			UrlUser:  j.Get(2),
		}
		listFrieds = append(listFrieds, temp)
	}
	return &listFrieds, nil
}
