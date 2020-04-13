package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

type config struct {
	BotToken   string `mapstructure:"bot-token"`
	ShardCount int    `mapstructure:"shard-count"`
	Author     struct {
		Name string `mapstructure:"name"`
		URL  string `mapstructure:"url"`
		Icon string `mapstructure:"icon"`
	} `mapstructure:"author"`
}

var cfg config

var dgs []*discordgo.Session

func init() {
	viper.AddConfigPath(".")
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Panic(err)
	}
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Panic(err)
	}

	initDb()
}

func main() {
	var err error
	dgs = make([]*discordgo.Session, 0)

	for i := 0; i < cfg.ShardCount; i++ {
		dg, err := discordgo.New("Bot " + cfg.BotToken)
		if err != nil {
			panic(err)
		}
		// Assign Shard
		dg.ShardID = i
		dg.ShardCount = cfg.ShardCount

		// Register the messageCreate func as a callback for MessageCreate events.
		dg.AddHandler(messageCreate)
		dgs = append(dgs, dg)
	}

	// Open a websocket connection to Discord and begin listening.
	for i, dg := range dgs {
		err = dg.Open()
		if err != nil {
			fmt.Printf("error opening connection for shard %v, %s", i, err.Error())
			return
		}
	}

	c := cron.New()
	c.AddFunc("CRON_TZ=Asia/Bangkok 00 12 * * *", broadcast)
	c.Start()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	go func() {
		broadcasted,err := getTodayBroadcastStatus()
		if err !=nil {
			fmt.Printf("Error getting today broadcast status, skipping.\n")
			return
		}
		if !broadcasted {
			broadcast()
		}
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	c.Stop()
	for _, dg := range dgs {
		dg.Close()
	}
	db.Close()
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
