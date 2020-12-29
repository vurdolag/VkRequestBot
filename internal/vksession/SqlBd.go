package vksession

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"sync"
)

const path string = "bd/vk.db"

type BdResponse struct {
	Id    string
	Count int
}

type DataBase struct {
	alreadyAddUser *BD
	maxAnswer      *BD
	newRepost      *BD
	isEnglish      *BD
}

func InitDataBase(muGlobal *sync.Mutex) *DataBase {
	return &DataBase{
		&BD{"already_add_user", muGlobal},
		&BD{"max_answer", muGlobal},
		&BD{"new_repost", muGlobal},
		&BD{"is_english", muGlobal},
	}
}

type BD struct {
	name     string
	muGlobal *sync.Mutex
}

func (self *BD) get(idUser string) (BdResponse, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return BdResponse{}, err
	}
	defer db.Close()
	q := fmt.Sprintf("select * from %s where id = \"%s\";", self.name, idUser)

	self.muGlobal.Lock()
	rows, err := db.Query(q)
	self.muGlobal.Unlock()

	if err != nil {
		return BdResponse{}, err
	}
	defer rows.Close()
	p := BdResponse{}

	lenRow, err := rows.Columns()

	if len(lenRow) == 2 {
		for rows.Next() {
			err = rows.Scan(&p.Id, &p.Count)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	} else if len(lenRow) == 1 {
		for rows.Next() {
			err = rows.Scan(&p.Id)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	} else {
		return BdResponse{}, nil
	}

	return p, nil
}

func (self *BD) put(idUser string, count int) (sql.Result, error) {
	var val string
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if count != -1 {
		val = fmt.Sprintf("\"%s\", %d", idUser, count)
	} else {
		val = fmt.Sprintf("\"%s\"", idUser)
	}

	q := fmt.Sprintf("insert into %s values (%s);", self.name, val)

	self.muGlobal.Lock()
	result, err := db.Exec(q)
	self.muGlobal.Unlock()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (self *BD) up(idUser string, count int) (sql.Result, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	q := fmt.Sprintf("update %s set count = %d where id = \"%s\";", self.name, count, idUser)

	self.muGlobal.Lock()
	result, err := db.Exec(q)
	self.muGlobal.Unlock()

	if err != nil {
		return nil, err
	}

	return result, nil
}
