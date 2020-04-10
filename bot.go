package main

import (
	"fmt"
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
			}
			t, err := time.Parse("02/01/2006 15:04", data.UpdateDate)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "เกิดข้อผิดพลาด กรุณาลองใหม่ภายหลัง")
			}
			embed := discordgo.MessageEmbed{
				Title: "รายงานสถานการณ์ โควิด-19",
				Author: &discordgo.MessageEmbedAuthor{
					Name:    cfg.Author.Name,
					IconURL: cfg.Author.Icon,
					URL:     cfg.Author.URL,
				},
				Color: 16721136,
				Provider: &discordgo.MessageEmbedProvider{
					Name: "กรมควบคุมโรค",
					URL:  "http://covid19.ddc.moph.go.th/",
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   fmt.Sprintf("%s", currentDateTH(t)),
						Value:  "\u200B",
						Inline: false,
					},
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
					Text: "ข้อมูลโดยกรมควบคุมโรค https://covid19.ddc.moph.go.th/",
				},
			}

			s.ChannelMessageSendEmbed(m.ChannelID, &embed)
		}
	}
}

func currentDateTH(t time.Time) string {
	d := days[int(t.Weekday())]
	m := months[int(t.Month())-1]

	return fmt.Sprintf("วัน%sที่ %v %s %v", d, t.Day(), m, t.Year()+543)
}
