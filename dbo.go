package main

import (
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type (
	channel struct {
		gorm.Model
		DiscordID string
		Active    bool
	}

	broadcastStamp struct {
		BroadcastDate string
	}
)

var db *gorm.DB

func initDb() {
	err := touchFile("data.db")
	if err != nil {
		log.Panic(err)
	}
	db, err = gorm.Open("sqlite3", "data.db")
	if err != nil {
		log.Panic(err)
	}
	db.AutoMigrate(channel{})
	db.AutoMigrate(broadcastStamp{})
}

func getSubs() (*[]channel, error) {
	chList := []channel{}
	err := db.Where(&channel{Active: true}).Find(&chList).Error

	if err != nil {
		return nil, err
	}

	return &chList, nil
}

func subscribe(channelID string) (bool, error) {
	ch := channel{}
	err := db.Where(channel{DiscordID: channelID}).First(&ch).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	if ch.ID == 0 {
		ch = channel{
			DiscordID: channelID,
			Active:    true,
		}
		err = db.Create(&ch).Error
		return true, nil
	}

	if ch.Active {
		return false, nil
	}
	ch.Active = true
	err = db.Save(&ch).Error

	if err != nil {
		return false, err
	}
	return true, nil

}

func unsubscribe(channelID string) (bool, error) {
	ch := channel{}
	err := db.Where(channel{DiscordID: channelID}).First(&ch).Error
	if err != nil {
		return false, err
	}

	if ch.ID == 0 || ch.Active == false {
		return false, nil
	}

	ch.Active = false

	err = db.Save(&ch).Error
	if err != nil {
		return false, err
	}

	return true, nil
}

func ubsubscribeBulk(chs []string) error {
	err := db.Table("channels").Where("discord_id in ?", chs).Updates(map[string]interface{}{"active": false}).Error
	return err
}

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

func getTodayBroadcastStatus() (bool, error) {
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	str := now.Format(time.RFC3339)
	c := 0
	err := db.Model(broadcastStamp{}).Where(&broadcastStamp{BroadcastDate: str}).Count(&c).Error
	if err != nil {
		return true, err
	}
	if c == 0 {
		return false, nil
	}
	return true, nil
}

func stampBroadcastDate() error {
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	str := now.Format(time.RFC3339)
	b := broadcastStamp{
		BroadcastDate: str,
	}
	err := db.Save(&b).Error
	if err != nil {
		return err
	}

	return nil
}
