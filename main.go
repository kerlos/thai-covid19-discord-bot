package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/robfig/cron/v3"
	"github.com/servusdei2018/shards"
	"github.com/spf13/viper"
)

type config struct {
	BotToken       string `mapstructure:"bot-token"`
	ShardCount     int    `mapstructure:"shard-count"`
	OwnerChannelID string `mapstructure:"owner-channel-id"`
	Author         struct {
		Name string `mapstructure:"name"`
		URL  string `mapstructure:"url"`
		Icon string `mapstructure:"icon"`
	} `mapstructure:"author"`
	BroadcastCron string `mapstructure:"broadcast-cron"`
}

var cfg config
var ca *cache.Cache
var loc *time.Location
var Mgr *shards.Manager

func init() {
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	ca = cache.New(30*time.Minute, 60*time.Minute)
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Panic(err)
	}
	loc, err = time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Panic(err)
	}
	initDb()
	//initCache()

	checkContent, err := os.ReadFile("covid19_check_result.json")
	if err != nil {
		log.Panic(err)
	}

	provinceContent, err := os.ReadFile("provinces.json")
	if err != nil {
		log.Panic(err)
	}

	var provincesData []province
	err = json.Unmarshal(provinceContent, &provincesData)
	if err != nil {
		log.Panic(err)
	}
	provinces = make(map[string]string)
	//most used province
	provinces["bkk"] = "bangkok"
	provinces["กรุงเทพ"] = "bangkok"
	provinces["กทม"] = "bangkok"
	for _, v := range provincesData {
		provinces[v.Slug] = v.Slug
		provinces[v.Title] = v.Slug
	}
	err = json.Unmarshal(checkContent, &checkResults)
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	var err error

	Mgr, err = shards.New("Bot " + cfg.BotToken)
	if err != nil {
		fmt.Println("[ERROR] Error creating manager,", err)
		return
	}
	Mgr.AddHandler(messageCreate)

	err = Mgr.Start()
	if err != nil {
		fmt.Println("[ERROR] Error starting manager,", err)
		return
	}
	Mgr.Shards[0].AddHandler(checkReactionAdd)
	Mgr.Shards[0].AddHandler(checkReactionRemove)
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	//broadcastSubs()

	c := cron.New()
	c.AddFunc(fmt.Sprintf("CRON_TZ=%s", cfg.BroadcastCron), broadcast)
	c.Start()
	/*
		now := time.Now().In(loc)
		noon := time.Date(now.Year(), now.Month(), now.Day(), 11, 59, 59, 0, loc)
		if now.After(noon) {
			broadcasted, err := getTodayBroadcastStatus()
			if err != nil {
				fmt.Printf("Error getting today broadcast status, skipping.\n")
			}
			if !broadcasted {
				broadcast()
			}
		}*/
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	c.Stop()
	db.Close()
	fmt.Println("[INFO] Stopping shard manager...")
	Mgr.Shutdown()
}

func broadcast() {
	retryCount := 0
	for {
		err := broadcastSubs()
		if err != nil {
			fmt.Printf("Error cron %s\n", err.Error())
			retryCount++
			if retryCount > 5 {
				break
			}
			time.Sleep(5 * time.Minute)
			continue
		}
		break
	}
}
