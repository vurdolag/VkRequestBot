package vk

type UserInfo struct {
	Response []UserInfoFields
}

type UserInfoFields struct {
	Id                int
	First_name        string
	Last_name         string
	Is_closed         bool
	Can_access_closed bool
	Deactivated       string
	Photo_200         string
	Sex               int
	Bdate             string
	Last_seen         LastSeen
}

type LastSeen struct {
	Time     int
	Platform int
}

type LongPollInitFields struct {
	Key    string
	Server string
	Ts     string
}

type Obj struct {
	Message     Msg
	Client_info ClientInfo
}

type Msg struct {
	Id                      int
	From_id                 int
	Owner_id                int
	Date                    int
	Peer_id                 int
	Text                    string
	Conversation_message_id int
	Important               bool
	Random_id               int
	Is_hidden               bool
	Attachments             []Atta
}

type ClientInfo struct {
	Button_actions  []string
	Keyboard        bool
	Lang_id         int
	Carousel        bool
	Inline_keyboard bool
}

type SizesPhoto struct {
	Height int
	Width  int
	Type   string
	Url    string
}

type PhotoAtta struct {
	Album_id   int
	Date       int
	Id         int
	Owner_id   int
	Has_tags   bool
	Access_key string
	Text       string
	Sizes      []SizesPhoto
}

type AudioMsgAtta struct {
	Id         int
	Owner_id   int
	Duration   int
	Waveform   []int
	Link_ogg   string
	Link_mp3   string
	Access_key string
}

type DocAtta struct {
	Id       int
	Owner_id int
	Date     int
	Title    string
	Size     int
	Ext      string
	Type     int
	Url      string
}

type Atta struct {
	Type          string
	Photo         PhotoAtta
	Audio_message AudioMsgAtta
	Doc           DocAtta
}

type UpdateLongPollFeed struct {
	Ts     string
	Events []string
}

type LongPollFields struct {
	Ts      int
	Updates [][]interface{}
	Failed  int
}

type PhotoUploadFields struct {
	Upload_url string
	Album_id   int
	Group_id   int
	User_id    int
}

type PhotoUplodAndSave struct {
	Server int
	Photo  string
	Hash   string
}

type ShortLinkFields struct {
	Short_url string
}

type KeyVal struct {
	Key string
	Val string
}

type ErrorFields struct {
	Method         string
	Error_code     int
	Error_msg      string
	Request_params []KeyVal
}

type TranslateResponse struct {
	Text []string
	Code int
	Lang string
}

type CheckTextFields struct {
	Pos  int
	Code int
	S    []string
	Len  int
}

type YandexGetIamTokenFields struct {
	IamToken string
}

type ModerationFields struct {
	Results []struct {
		Results []struct {
			Classification struct {
				Properties []struct {
					Name        string
					Probability float32
				}
			}
		}
	}
}

type Payload struct {
	Payload []interface{}
}

type Requests struct {
	Requests [][]interface{}
}

type OutRequests struct {
	Out_requests [][]interface{}
}

type ImInit struct {
	Tabs map[string]map[string]interface{}
}

type Group struct {
	Id        string
	Name      string
	UrlGroup  string
	UrlPhoto  string
	HashLeave string
}

type User struct {
	Id       string
	Name     string
	UrlPhoto string
	UrlUser  string
}

type FriendAll struct {
	All [][]string
}

type FeedUpdate struct {
	Version   int
	Type      string
	Link      string
	Text      string
	Author_id int
	Title     string
}

type Akk struct {
	Name      string
	Id        string
	Login     string
	Password  string
	Useragent string
	Proxy     string
}
