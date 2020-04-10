package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

type config struct {
	BotToken string `mapstructure:"bot-token"`
	Author   struct {
		Name string `mapstructure:"name"`
		URL  string `mapstructure:"url"`
		Icon string `mapstructure:"icon"`
	} `mapstructure:"author"`
}

var cfg config
var dg *discordgo.Session

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
	dg, err = discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	c := cron.New()
	c.AddFunc("CRON_TZ=Asia/Bangkok 00 12 * * *", func() {
		err := broadcastSubs()
		if err != nil {
			fmt.Printf("Error cron %s\n", err.Error())
		}
	})
	c.Start()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	c.Stop()
	dg.Close()
	db.Close()
}
