package vksession

import (
	"VkRequestBot/configs"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/buger/jsonparser"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const urlVk = "https://api.vk.com/method/"
const vApi = "5.122"

const im = "https://vk.com/al_im.php"
const al_groups = "https://vk.com/al_groups.php"
const al_public = "https://vk.com/al_public.php"
const like = "https://vk.com/like.php"
const hints = "https://vk.com/hints.php"
const al_friends = "https://vk.com/al_friends.php"
const al_wall = "https://vk.com/al_wall.php"
const al_photos = "https://vk.com/al_photos.php"
const al_page = "https://vk.com/al_page.php"
const friends = "https://vk.com/friends"
const al_settings = "https://vk.com/al_settings.php"

var moderToken string
var lastUpdateModerToken int

func requestsGet(url_get string, params map[string]string) ([]byte, error) {
	var par string

	for key, val := range params {
		par += key + "=" + val + "&"
	}

	par = url_get + "?" + par

	client := http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(par)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)

	runtime.GC()
	return body, nil
}

func postFileFromByte(targetUrl string, params map[string]string, byteData []byte) ([]byte, error) {
	buf := bytes.NewReader(byteData)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "file.jpg")
	if err != nil {
		return nil, err
	}

	_, _ = io.Copy(part, buf)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	_ = writer.Close()
	req, err := http.NewRequest("POST", targetUrl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return resp_body, nil
}

func postFileFromJson(targetUrl string, params map[string]string, byteData []byte) ([]byte, error) {
	buf := bytes.NewReader(byteData)

	var par string

	for key, val := range params {
		par += key + "=" + val + "&"
	}

	targetUrl = targetUrl + "?" + par

	req, err := http.NewRequest("POST", targetUrl, buf)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	runtime.GC()
	return resp_body, nil
}

func requestsPost(targetUrl string, params, headers map[string]string, byteData []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewReader(byteData))
	if err != nil {
		return nil, err
	}

	if params != nil {
		for key, val := range params {
			req.Form.Set(key, val)
		}
	}

	if headers != nil {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	runtime.GC()
	return respBody, nil
}

func postForm(targetUrl string, form url.Values) ([]byte, error) {
	resp, err := http.PostForm(targetUrl, form)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	runtime.GC()
	return body, nil
}

func LoadAkks(path string) map[int]string {
	file, err := os.Open(path) //"files/data.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()
	c, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var s map[int]string
	json.Unmarshal(c, &s)
	return s

}

func LoadFile(path string) ([]byte, error) {
	r, err := loadFile(path)
	return r, err
}

func loadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	c, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	runtime.GC()
	return c, nil
}

func WriteFile(path, source string) bool {
	return writeNewFileTxt(path, source)
}

func writeNewFileTxt(path, source string) bool {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Unable to create file:", err)
		return false
	}
	defer file.Close()
	_, err = file.Write([]byte(source))
	if err != nil {
		fmt.Println(err)
	}
	return true
}

func translate(text string, lang string) string {
	params := map[string]string{
		"key":  configs.YANDEX_TRANSLATE_TOKEN,
		"text": url.QueryEscape(text),
		"lang": lang,
	}

	res, err := requestsGet("https://translate.yandex.net/api/v1.5/tr.json/translate", params)

	fmt.Println(text, lang)

	fmt.Println(string(res))

	if err != nil {
		fmt.Println(err)
		return "Ошибка перевода =("
	}

	var response TranslateResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		fmt.Println(err)
		return "Ошибка перевода =("
	}

	return response.Text[0]
}

func checkText(text string) string {
	params := map[string]string{"text": url.QueryEscape(text)}

	resp, err := requestsGet("https://speller.yandex.net/services/spellservice.json/checkText", params)
	if err != nil {
		fmt.Println("error checkText", err)
		return text
	}

	fmt.Println(string(resp))

	var result []CheckTextFields
	err = json.Unmarshal(resp, &result)
	if err != nil {
		fmt.Println("error checkText json", err)
		return text
	}

	if result != nil {
		ind := 0
		t := []rune(text)
		var tmp []rune

		for _, word := range result {
			var tt []rune
			tt = t[ind:word.Pos]
			tt = append(tt, []rune(word.S[0])...)
			tmp = append(tmp, tt...)
			ind = word.Pos + word.Len
		}

		tmp = append(tmp, t[ind:]...)

		fmt.Println("go text", string(tmp))

		return string(tmp)

	} else {
		fmt.Println("origin text")

		return text
	}
}

func yandexGetIamToken() string {
	params := map[string]string{
		"yandexPassportOauthToken": configs.YANDEX_MODERATION_TOKEN,
	}
	body, err := postFileFromJson("https://iam.api.cloud.yandex.net/iam/v1/tokens", params, nil)

	if err != nil {
		fmt.Println("err", err)
		return ""
	}

	var resp YandexGetIamTokenFields
	_ = json.Unmarshal(body, &resp)

	return resp.IamToken
}

func moderationImg(img []byte) (map[string]float32, error) {
	data := `{"folderId": "` + configs.YANDEX_FOLDER_ID + `","analyze_specs": [{"content":"` +
		base64.StdEncoding.EncodeToString(img) +
		`","features": [{"type": "CLASSIFICATION","classificationConfig": {"model": "moderation"}}]}]}`

	if moderToken == "" || time.Now().Second()-lastUpdateModerToken > 3600*6 {
		lastUpdateModerToken = time.Now().Second()
		moderToken = yandexGetIamToken()
	}

	head := map[string]string{
		"Authorization": "Bearer " + moderToken,
	}

	resp, err := requestsPost("https://vision.api.cloud.yandex.net/vision/v1/batchAnalyze", nil, head, []byte(data))
	if err != nil {
		return nil, err
	}

	var results ModerationFields

	err = json.Unmarshal(resp, &results)
	if err != nil {
		return nil, err
	}

	response := make(map[string]float32)

	if len(results.Results) == 0 {
		return nil, nil
	}

	for _, val := range results.Results[0].Results[0].Classification.Properties {
		response[val.Name] = val.Probability
	}

	runtime.GC()
	return response, nil
}

func RandSleep(randTime, addTime int) {
	randSleep(randTime, addTime)
}

func randSleep(randTime, addTime int) {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(randTime*100) + 1
	t := time.Second/100*time.Duration(r) + time.Second*time.Duration(addTime)
	time.Sleep(t)
	return
}

func randChoice(list []string) string {
	rand.Seed(time.Now().UnixNano())
	return list[rand.Intn(len(list))]
}

func logs(str ...string) bool {
	content := ""
	name := "log"

	if len(str) == 1 {
		content = str[0]
	} else if len(str) > 1 {
		content = str[0]
		name = str[1]
	}

	content = content + "\n"

	file, err := os.OpenFile("logs/"+name+".txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println(err)
		writeNewFileTxt("logs/"+name+".txt", content)
		return false
	}
	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func logsErr(err error) {
	t := time.Now().Unix()
	logs(fm("%d > %v", t, err), "error")
	fmt.Println(err)
}

func IsIn(str string, list []string) bool {
	if list != nil {
		return false
	}

	for index := range list {
		if str == list[index] {
			return true
		}
	}
	return false
}

func IsIntInList(str int, list []int) bool {
	for index := range list {
		if str == list[index] {
			return true
		}
	}
	return false
}

func st(i int) string {
	return strconv.Itoa(i)
}

func isRand(v float32) bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Float32()*100 < v
}

type RE struct {
	compile  map[string]*regexp.Regexp
	muGlobal *sync.Mutex
}

func InitRE(muGlobal *sync.Mutex) *RE {
	return &RE{
		compile:  make(map[string]*regexp.Regexp, 1000),
		muGlobal: muGlobal,
	}
}

func (self *RE) FindAllBetween(Str *[]byte, str1, str2 string) []string {
	sourceStr := string(*Str)

	var ok bool
	var err error
	var str, x string
	var finalEditStrings = make([]string, 0, 5)
	var findStrings = make([]string, 0, 5)
	var lenStr1, lenStr2, lenX int
	var compiledString *regexp.Regexp

	lenStr1 = len(reSubBackSlash.ReplaceAllString(str1, ""))
	lenStr2 = len(reSubBackSlash.ReplaceAllString(str2, ""))

	str = fmt.Sprintf("%s.+?%s", str1, str2)

	self.muGlobal.Lock()
	compiledString, ok = self.compile[str]
	self.muGlobal.Unlock()

	if !ok {
		compiledString, err = regexp.Compile(str)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 10)
		}
		self.muGlobal.Lock()
		self.compile[str] = compiledString
		self.muGlobal.Unlock()
	}

	findStrings = compiledString.FindAllString(sourceStr, -1)

	for _, val := range findStrings {
		lenX = len(val)
		x = val[lenStr1 : lenX-lenStr2]
		finalEditStrings = append(finalEditStrings, x)
	}

	runtime.GC()
	return finalEditStrings
}

func (self *RE) FindAllBetweenStr(sourceStr, str1, str2 string) []string {
	var ok bool
	var err error
	var str, x string
	var finalEditStrings = make([]string, 0, 5)
	var findStrings = make([]string, 0, 5)
	var lenStr1, lenStr2, lenX int
	var compiledString *regexp.Regexp

	lenStr1 = len(reSubBackSlash.ReplaceAllString(str1, ""))
	lenStr2 = len(reSubBackSlash.ReplaceAllString(str2, ""))

	str = fmt.Sprintf("%s.+?%s", str1, str2)

	self.muGlobal.Lock()
	compiledString, ok = self.compile[str]
	self.muGlobal.Unlock()

	if !ok {
		compiledString, err = regexp.Compile(str)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second * 10)
		}

		self.muGlobal.Lock()
		self.compile[str] = compiledString
		self.muGlobal.Unlock()
	}

	findStrings = compiledString.FindAllString(sourceStr, -1)

	for _, val := range findStrings {
		lenX = len(val)
		x = val[lenStr1 : lenX-lenStr2]
		finalEditStrings = append(finalEditStrings, x)
	}

	runtime.GC()
	return finalEditStrings
}

func (self *RE) Check(sourceStr, pattern string) bool {
	var compiledString *regexp.Regexp
	var ok bool
	var err error

	self.muGlobal.Lock()
	compiledString, ok = self.compile[pattern]
	self.muGlobal.Unlock()

	if !ok {
		compiledString, err = regexp.Compile(pattern)
		if err != nil {
			fmt.Println("error in Check", err, sourceStr, pattern)
			time.Sleep(time.Second * 1)
			return false
		}
		self.muGlobal.Lock()
		self.compile[pattern] = compiledString
		self.muGlobal.Unlock()
	}
	return compiledString.MatchString(sourceStr)
}

func getUserInfoFromApi(userIds ...string) (*[]UserInfoFields, error) {
	var fin []UserInfoFields
	var users []string

	for x := 0; x < len(userIds); x += 1000 {
		users = userIds[x:]
		if len(users) > 1000 {
			users = users[:1000]
		}

		form := url.Values{
			"v":            {vApi},
			"access_token": {configs.TOKEN_GROUP},
			"fields":       {"photo_200,last_seen"},
			"user_ids":     {strings.Join(users, ",")},
		}
		res, err := postForm(urlVk+"users.get", form)
		if err != nil {
			return &[]UserInfoFields{}, err
		}

		var response UserInfo
		err = json.Unmarshal(res, &response)
		if err != nil || len(response.Response) == 0 {
			return &[]UserInfoFields{}, err
		}

		for index := range response.Response {
			fin = append(fin, response.Response[index])
		}
		time.Sleep(time.Millisecond * 100)
	}
	return &fin, nil
}

func jsInt(b []byte, q ...string) (int, error) {
	var err error
	var s int64
	if len(q) != 0 {
		s, err = jsonparser.GetInt(b, strings.Split(q[0], ".")...)
	} else {
		s, err = jsonparser.GetInt(b)
	}
	return int(s), err
}

func jsStr(b []byte, q ...string) (string, error) {
	var err error
	var s string
	if len(q) != 0 {
		s, err = jsonparser.GetString(b, strings.Split(q[0], ".")...)
	} else {
		s, err = jsonparser.GetString(b)
	}
	return s, err
}

func jsArr(b []byte, q ...string) ([][]byte, error) {
	v := make([][]byte, 0, 15)
	handler := func(value []byte, dataType jsonparser.ValueType, offset int, err error) { v = append(v, value) }
	var err error
	if len(q) != 0 {
		_, err = jsonparser.ArrayEach(b, handler, strings.Split(q[0], ".")...)
	} else {
		_, err = jsonparser.ArrayEach(b, handler)
	}
	return v, err
}

func jsObj(b []byte, q ...string) (map[string][]byte, error) {
	v := make(map[string][]byte, 30)
	handler := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		v[string(key)] = value
		return nil
	}
	var err error
	if len(q) != 0 {
		err = jsonparser.ObjectEach(b, handler, strings.Split(q[0], ".")...)
	} else {
		err = jsonparser.ObjectEach(b, handler)
	}
	return v, err
}

func js(b []byte, q ...string) (string, error) {
	var err error
	var s []byte
	if len(q) != 0 {
		s, _, _, err = jsonparser.Get(b, strings.Split(q[0], ".")...)
	} else {
		s, _, _, err = jsonparser.Get(b)
	}
	return string(s), err
}

func jsAsByte(b []byte, q ...string) ([]byte, error) {
	var err error
	var s []byte
	if len(q) != 0 {
		s, _, _, err = jsonparser.Get(b, strings.Split(q[0], ".")...)
	} else {
		s, _, _, err = jsonparser.Get(b)
	}
	return s, err
}

func delBackSlash(b []byte) []byte {
	p := []byte(`\`)[0]
	t := make([]byte, 0, len(b))
	sw := false
	for i := range b {
		if b[i] == p && !sw {
			sw = true
		} else if b[i] != p {
			sw = false
			t = append(t, b[i])
		}
	}
	return t
}

func dBackSlash(b []byte) []byte {
	p := []byte(`\`)[0]
	t := make([]byte, 0, len(b))
	for i := range b {
		if (b[i] == p && b[i+1] == p) || b[i] != p {
			t = append(t, b[i])
		}
	}
	return t
}

func fm(f string, v ...interface{}) string {
	//форматирование
	return fmt.Sprintf(f, v...)
}

type Js struct {
	data []byte
}

func NewJs(b []byte) *Js {
	return &Js{b}
}

func (j *Js) Get(q string) string {
	r, err := jsAsByte(j.data, q)
	if err != nil {
		if configs.DEBUG {
			logsErr(err)
		}
		return ""
	}
	return string(delBackSlash(r))
}

func (j *Js) Int(q string) int {
	r, err := jsInt(j.data, q)
	if err != nil {
		if configs.DEBUG {
			logsErr(err)
		}
		return 0
	}
	return r
}

type JsArray struct {
	data []byte
	len  int
}

func NewJsArray(b []byte) *JsArray {
	l, _ := jsArr(b)
	return &JsArray{b, len(l)}
}

func (j *JsArray) Get(index int) string {
	r, err := jsAsByte(j.data, fm("[%d]", index))
	if err != nil {
		if configs.DEBUG {
			logsErr(err)
		}
		return ""
	}
	return string(delBackSlash(r))
}

func (j *JsArray) Int(index int) int {
	r, err := jsInt(j.data, fm("[%d]", index))
	if err != nil {
		if configs.DEBUG {
			logsErr(err)
		}
		return 0
	}
	return r
}

var ruDict = []rune("ёйцукенгшщзхъфывапролджэячсмитьбю")
var enDict = []rune("QWERTYUIOPASDFGHJKLZXCVBNMqertyui")

func enc(v string) string {
	w := make([]rune, 0, 100)
	q := []rune(v)
	sw := false
	in := 0
	for _, i := range q {
		sw = false
		for ind, x := range ruDict {
			in = ind
			if i == x {
				sw = true
				break
			}
		}
		if sw {
			w = append(w, enDict[in])
		} else {
			w = append(w, i)
		}
	}
	return string(w)
}

func dec(v string) string {
	w := make([]rune, 0, 100)
	q := []rune(v)
	sw := false
	in := 0
	for _, i := range q {
		sw = false
		for ind, x := range enDict {
			in = ind
			if i == x {
				sw = true
				break
			}
		}
		if sw {
			w = append(w, ruDict[in])
		} else {
			w = append(w, i)
		}
	}
	return string(w)
}

func finder(b []byte, str1, str2 string, count int) []string {
	r := make([]string, 0, 15)
	var a, y, i, c int

	s1 := []byte(str1)
	s2 := []byte(str2)

	a = -1
	switcher := true

	for i <= len(b)-1 {
		if y == 0 && b[i] == s1[0] && switcher {
			if len(s1) == 1 {
				a = i
				switcher = false
				i++
			} else {
				for x := 1; x < len(s1); x++ {
					i++
					if b[i] != s1[x] {
						a = -1
						break
					}
					if x == len(s1)-1 {
						a = i
						switcher = false
						i++
					}
				}
			}
		}
		if a != -1 && b[i] == s2[0] {
			if len(s2) == 1 {
				y = i - 1
				switcher = true
			} else {
				for x := 1; x < len(s2); x++ {
					i++
					if b[i] != s2[x] {
						y = 0
						break
					}
					if x == len(s2)-1 {
						y = i - len(s2)
						switcher = true
					}
				}
			}
		}
		if y != 0 {
			r = append(r, string(b[a+1:y+1]))
			a, y = -1, 0
			c++
			if c == count {
				break
			}
		}
		i++
	}
	return r
}

func finderFirst(b []byte, str1, str2 string) string {
	r := finder(b, str1, str2, 1)
	if len(r) == 0 {
		return ""
	}
	return r[0]

}
