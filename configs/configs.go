package configs

import (
	"github.com/joho/godotenv"
	"os"
)

const DEBUG = false

var C ConfI

type ConfI interface {
	GetPort() string
	GetAccountPath() string
	GetCookiePath() string
	GetAnswerPath() string
	YaTokenTranslate() string
	YaTokenModeration() string
	YaFolder() string
	Token() string
}

type Conf struct {
	port        string
	cookiePath  string
	accountPath string
	answerPath  string
	yaTokenT    string
	yaTokenM    string
	yaFolder    string
	token       string
}

func (c *Conf) GetPort() string {
	return c.port
}

func (c *Conf) GetAccountPath() string {
	return c.accountPath
}

func (c *Conf) GetCookiePath() string {
	return c.cookiePath
}

func (c *Conf) GetAnswerPath() string {
	return c.answerPath
}

func (c *Conf) YaTokenTranslate() string {
	return c.yaTokenT
}

func (c *Conf) YaTokenModeration() string {
	return c.yaTokenM
}

func (c *Conf) YaFolder() string {
	return c.yaFolder
}

func (c *Conf) Token() string {
	return c.token
}

func New(path string) (ConfI, error) {
	err := godotenv.Load(path)
	if err != nil {
		return nil, err
	}
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		panic("empty port")
	}
	accountPath := os.Getenv("ACCOUNT_PATH")
	cookiePath := os.Getenv("COOKIE_PATH")
	answerPath := os.Getenv("ANSWER_PATH")
	yaTokenT := os.Getenv("YA_TOKEN_TRANSLATE")
	yaTokenM := os.Getenv("YA_TOKEN_MODERATION")
	yaFolder := os.Getenv("YA_FOLDER")
	token := os.Getenv("TOKEN_GROUP")

	C = &Conf{port: port,
		accountPath: accountPath,
		cookiePath:  cookiePath,
		answerPath:  answerPath,
		yaTokenT:    yaTokenT,
		yaFolder:    yaFolder,
		yaTokenM:    yaTokenM,
		token:       token,
	}
	return C, nil
}
