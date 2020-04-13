package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
)

var days = []string{
	"อาทิตย์",
	"จันทร์",
	"อังคาร",
	"พุธ",
	"พฤหัสบดี",
	"ศุกร์",
	"เสาร์",
}

var months = []string{
	"มกราคม",
	"กุมภาพันธ์",
	"มีนาคม",
	"เมษายน",
	"พฤษภาคม",
	"มิถุนายน",
	"กรกฎาคม",
	"สิงหาคม",
	"กันยายน",
	"ตุลาคม",
	"พฤษจิกายน",
	"ธันวาคม",
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.ToLower(m.Content)

	if strings.HasPrefix(content, "/covid") {
		prms := strings.Split(content, " ")
		if len(prms) == 1 || prms[1] == "today" {
			data, err := getData()
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "เกิดข้อผิดพลาด กรุณาลองใหม่ภายหลัง")
				return
			}
			embed, err := buildEmbed(data)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "เกิดข้อผิดพลาด กรุณาลองใหม่ภายหลัง")
				return
			}
			s.ChannelMessageSendEmbed(m.ChannelID, embed)
		}

		if len(prms) > 1 {
			switch prms[1] {
			case "sub", "subscribe":
				_, err := subscribe(m.ChannelID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "เกิดข้อผิดพลาด กรุณาลองใหม่ภายหลัง")
					return
				}
				/*
					if !ok {
						s.ChannelMessageSend(m.ChannelID, "ช่องนี้ได้ติดตามข่าวอยู่แล้ว")
					}*/

				s.ChannelMessageSend(m.ChannelID, "ติดตามข่าวเรียบร้อย")
				break

			case "unsub", "unsubscribe":
				_, err := unsubscribe(m.ChannelID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "เกิดข้อผิดพลาด กรุณาลองใหม่ภายหลัง")
					return
				}
				/*
					if !ok {
						s.ChannelMessageSend(m.ChannelID, "ช่องนี้ไม่ได้ติดตามข่าว")
					}*/

				s.ChannelMessageSend(m.ChannelID, "ยกเลิกการติดตามข่าวเรียบร้อย")
				break
			case "help":
				s.ChannelMessageSend(m.ChannelID, "พิมพ์ `/covid` เพื่อดูรายงานปัจจุบัน\nพิมพ์ `/covid sub` เพื่อติดตามข่าวอัตโนมัติทุกวันเวลา 12.00 น.\nพิมพ์ `/covid unsub` เพื่อยกเลิกการติดตามข่าว")
				break
			}
		}
	}
}

func broadcastSubs() error {
	chList, err := getSubs()
	if err != nil {
		return err
	}
	now := time.Now()
	var data *covidData
	delayNotice := true
	for {
		data, err = getData()
		if err != nil {
			return err
		}

		t, err := time.Parse("02/01/2006 15:04", data.UpdateDate)
		if err != nil {
			return err
		}

		if dateEqual(t, now) {
			break
		}
		fmt.Printf("broadcast data not update, retrying...\n")
		if delayNotice {
			for _, ch := range *chList {
				shardID := getShardID(ch.DiscordID)
				dgs[shardID].ChannelMessageSend(ch.DiscordID, "ข้อมูลจากกรมควบคุมโรคยังไม่มีข้อมูลใหม่ ระบบจะส่งข้อมูลอีกครั้งหลังจากได้รับข้อมูลใหม่แล้ว")
				time.Sleep(100 * time.Millisecond)
			}
			delayNotice = false
		}
		time.Sleep(5 * time.Minute)
	}

	embed, err := buildEmbed(data)
	if err != nil {
		return err
	}

	retriedList := make([]string, 0)
	retriedCount := 1

	for _, ch := range *chList {
		shardID := getShardID(ch.DiscordID)
		_, err = dgs[shardID].ChannelMessageSendEmbed(ch.DiscordID, embed)
		if err != nil {
			retriedList = append(retriedList, ch.DiscordID)
		}
		time.Sleep(100 * time.Millisecond)
	}

	for {
		if len(retriedList) > 0 {
			fmt.Printf("%v channel failed to deliver. retry attempted: %v\n", len(retriedList), retriedCount)
			if retriedCount > 3 {
				break
			}
			tmp := make([]string, 0)
			time.Sleep(1 * time.Minute)
			for _, id := range retriedList {
				_, err = dgs[0].ChannelMessageSendEmbed(id, embed)
				if err != nil {
					tmp = append(tmp, id)
				}
				time.Sleep(100 * time.Millisecond)
			}
			retriedList = tmp
			retriedCount++
		} else {
			break
		}
	}

	return nil
}

func currentDateTH(t time.Time) string {
	d := days[int(t.Weekday())]
	m := months[int(t.Month())-1]

	return fmt.Sprintf("วัน%sที่ %v %s %v", d, t.Day(), m, t.Year()+543)
}

func buildEmbed(data *covidData) (*discordgo.MessageEmbed, error) {
	t, err := time.Parse("02/01/2006 15:04", data.UpdateDate)
	if err != nil {
		return nil, err
	}
	embed := discordgo.MessageEmbed{
		Title: "รายงานสถานการณ์ โควิด-19 ในประเทศไทย",
		/*
			Author: &discordgo.MessageEmbedAuthor{
				Name:    cfg.Author.Name,
				IconURL: cfg.Author.Icon,
				URL:     cfg.Author.URL,
			},*/

		Description: fmt.Sprintf("%s", currentDateTH(t)),
		Color:       16721136,
		Provider: &discordgo.MessageEmbedProvider{
			Name: "กรมควบคุมโรค",
			URL:  "http://covid19.ddc.moph.go.th/",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🤒 ติดเชื้อสะสม",
				Value:  fmt.Sprintf("%s (เพิ่มขึ้น %s)", humanize.Comma(int64(data.Confirmed)), humanize.Comma(int64(data.NewConfirmed))),
				Inline: true,
			},
			{
				Name:   "💀 เสียชีวิต",
				Value:  fmt.Sprintf("%s (เพิ่มขึ้น %s)", humanize.Comma(int64(data.Deaths)), humanize.Comma(int64(data.NewDeaths))),
				Inline: true,
			},
			{
				Name:   "💪 หายแล้ว",
				Value:  fmt.Sprintf("%s (เพิ่มขึ้น %s)", humanize.Comma(int64(data.Recovered)), humanize.Comma(int64(data.NewRecovered))),
				Inline: true,
			},
			{
				Name:   "🏥 รักษาอยู่ใน รพ.",
				Value:  fmt.Sprintf("%s", humanize.Comma(int64(data.Hospitalized))),
				Inline: true,
			},
			{
				Name:   "🟥 อัตราการเสียชีวิต",
				Value:  fmt.Sprintf("%.2f%%", (float64(data.Deaths)/float64(data.Confirmed))*100),
				Inline: true,
			},
			{
				Name:   "🟩 อัตราการหาย",
				Value:  fmt.Sprintf("%.2f%%", (float64(data.Recovered)/float64(data.Confirmed))*100),
				Inline: true,
			},
		},
		URL: "https://covid19.ddc.moph.go.th/",
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("ข้อมูลโดย กรมควบคุมโรค\nบอทโดย %s\n%s", cfg.Author.Name, cfg.Author.URL),
		},
	}

	return &embed, nil
}

func dateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func getShardID(channelID string) int {
	if cfg.ShardCount == 1 {
		return 0
	}
	gid, err := strconv.ParseUint(channelID, 10, 64)
	if err != nil {
		return 0
	}
	shardID := (gid >> 22) % uint64(cfg.ShardCount)
	return int(shardID)
}
