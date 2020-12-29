package configs

import (
	"github.com/joho/godotenv"
	"os"
)

const YANDEX_TRANSLATE_TOKEN = "trnsl.1.1.20191026T085210Z.d618f941c09cba4c.64c40d3e7e87d85ff46276c7fce0b057e2ffd122"
const YANDEX_MODERATION_TOKEN = "AgAAAAADRT4KAATuwUCxpOqfKk18sXD7mZxzJEc"
const YANDEX_FOLDER_ID = "b1gb9k5cliueoqc4sg0t"

const TOKEN_SERVICE = "a8357cbda8357cbda8357cbdf6a85e531aaa835a8357cbdf5329f427fd9a1583ef604a9"
const TOKEN_GROUP = "6c30a85b9085b8dcb062fd0077f757ae24374e529ef4ee550dcf7f5b7ca7089038a4b4adaa9d4ae2d70a8"

const DEBUG = false

type ConfI interface {
	GetPort() string
	GetAccountPath() string
	GetCookiePath() string
	GetAnswerPath() string
}

type Conf struct {
	port        string
	cookiePath  string
	accountPath string
	answerPath  string
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

func New(path string) (*Conf, error) {
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

	return &Conf{port: port,
		accountPath: accountPath,
		cookiePath:  cookiePath,
		answerPath:  answerPath,
	}, nil
}
