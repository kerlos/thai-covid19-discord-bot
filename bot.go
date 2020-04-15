package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
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

const messageError = "เกิดข้อผิดพลาด กรุณาลองใหม่ภายหลัง"

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	content := strings.ToLower(m.Content)

	if strings.HasPrefix(content, "/covid") {
		prms := strings.Split(content, " ")
		if len(prms) == 1 || prms[1] == "today" {
			msgData := &discordgo.MessageSend{}
			var embed *discordgo.MessageEmbed
			var file *bytes.Buffer
			t := time.Now()
			t = t.In(loc)
			if embedCache, found := ca.Get("embed"); found {
				embed = embedCache.(*discordgo.MessageEmbed)
			}

			if embed == nil {
				data, err := getData()
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, messageError)
					return
				}
				td, err := time.Parse("02/01/2006 15:04", data.UpdateDate)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, messageError)
					return
				}
				embed, err = buildEmbed(data)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, messageError)
					return
				}
				if imgCache, ok := ca.Get(fmt.Sprintf("chart-%s", td.Format("Jan2"))); ok {
					embed.Image = imgCache.(*discordgo.MessageEmbedImage)
				}

				if embed.Image == nil {
					file, err = buildChart()
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, messageError)
						return
					}
					chart := &discordgo.File{
						Name:        fmt.Sprintf("covid-chart-%s.png", td.Format("20060102")),
						ContentType: "image/png",
						Reader:      file,
					}
					msgData.Files = append(msgData.Files, chart)
					t = td
				}
			}

			msgData.Embed = embed
			resp, err := s.ChannelMessageSendComplex(m.ChannelID, msgData)
			if err != nil {
				return
			}
			if embed.Image == nil {
				at := resp.Attachments[0]
				embed.Image = &discordgo.MessageEmbedImage{
					URL:      at.URL,
					ProxyURL: at.ProxyURL,
					Height:   at.Height,
					Width:    at.Width,
				}

				ca.Set(fmt.Sprintf("chart-%s", t.Format("Jan2")), embed.Image, 36*time.Hour)
				ca.Set("embed", embed, 30*time.Minute)
				s.ChannelMessageEditEmbed(m.ChannelID, resp.ID, embed)
			}
		}

		if len(prms) > 1 {
			switch prms[1] {
			case "sub", "subscribe":
				_, err := subscribe(m.ChannelID)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, messageError)
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
					s.ChannelMessageSend(m.ChannelID, messageError)
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
	now := time.Now().In(loc)
	var data *covidData
	//delayNotice := true
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
			now = t
			break
		}
		fmt.Printf("broadcast data not update, retrying...\n")
		time.Sleep(5 * time.Minute)

	}

	embed, err := buildEmbed(data)
	if err != nil {
		return err
	}

	file, err := buildChart()
	if err != nil {
		return err
	}

	chart := &discordgo.File{
		Name:        fmt.Sprintf("covid-chart-%s.png", now.Format("20060102")),
		ContentType: "image/png",
		Reader:      file,
	}
	msgData := &discordgo.MessageSend{
		Embed: embed,
	}
	msgData.Files = append(msgData.Files, chart)

	retriedList := make([]string, 0)
	retriedCount := 1

	for _, ch := range *chList {
		shardID := getShardID(ch.DiscordID)
		resp, err := dgs[shardID].ChannelMessageSendComplex(ch.DiscordID, msgData)
		if err != nil {
			retriedList = append(retriedList, ch.DiscordID)
		}

		if embed.Image == nil {
			at := resp.Attachments[0]
			embed.Image = &discordgo.MessageEmbedImage{
				URL:      at.URL,
				ProxyURL: at.ProxyURL,
				Height:   at.Height,
				Width:    at.Width,
			}
			ca.Set(fmt.Sprintf("chart-%s", now.Format("Jan2")), embed.Image, 36*time.Hour)
			ca.Set("embed", embed, 30*time.Minute)
			dgs[shardID].ChannelMessageEditEmbed(ch.DiscordID, resp.ID, embed)
		}
		time.Sleep(100 * time.Millisecond)
	}

	for {
		if len(retriedList) > 0 {
			if retriedCount > 3 {
				fmt.Printf("%v channels unsubscribe after 3 retries\n", len(retriedList))
				ubsubscribeBulk(retriedList)
				break
			}
			fmt.Printf("%v channel failed to deliver. retry attempted: %v\n", len(retriedList), retriedCount)
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
	now = time.Now().In(loc)
	fmt.Printf("finished broadcast at %s\n", now.Format(time.Stamp))
	err = stampBroadcastDate()
	if err != nil {
		fmt.Printf("error stamp broadcast date %s", err.Error())
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

func buildChart() (*bytes.Buffer, error) {
	data, err := getChartData()
	if err != nil {
		return nil, err
	}

	dlen := len(data.Data) - 30
	if dlen < 0 {
		dlen = 0
	}

	ttfData, err := ioutil.ReadFile("font/Kanit-Medium.ttf")
	if err != nil {
		log.Fatal(err)
	}
	f, err := truetype.Parse(ttfData)
	if err != nil {
		log.Fatal(err)
	}

	dt := data.Data[dlen:]
	dlen = len(dt)
	ticks := make([]chart.Tick, dlen)
	max := 0
	c := chart.ContinuousSeries{
		Name:    "ติดเชื้อสะสม",
		XValues: make([]float64, dlen),
		YValues: make([]float64, dlen),
		Style: chart.Style{
			StrokeColor: drawing.ColorFromHex("e1298e"),
			FillColor:   drawing.ColorFromHex("e1298e").WithAlpha(32),
			Show:        true,
		},
	}
	d := chart.ContinuousSeries{
		Name:    "เสียชีวิต",
		XValues: make([]float64, dlen),
		YValues: make([]float64, dlen),
		Style: chart.Style{
			StrokeColor: drawing.ColorBlack,
			FillColor:   drawing.ColorBlack.WithAlpha(32),
			Show:        true,
		},
	}
	r := chart.ContinuousSeries{
		Name:    "หายแล้ว",
		XValues: make([]float64, dlen),
		YValues: make([]float64, dlen),
		Style: chart.Style{
			StrokeColor: drawing.ColorFromHex("046034"),
			FillColor:   drawing.ColorFromHex("046034").WithAlpha(32),
			Show:        true,
		},
	}
	h := chart.ContinuousSeries{
		Name:    "รักษาอยู่ใน รพ.",
		XValues: make([]float64, dlen),
		YValues: make([]float64, dlen),
		Style: chart.Style{
			StrokeColor: drawing.ColorFromHex("179c9b"),
			FillColor:   drawing.ColorFromHex("179c9b").WithAlpha(32),
			Show:        true,
		},
	}
	for i, v := range dt {
		t, err := time.Parse("01/02/2006", dt[i].Date)
		if err != nil {
			return nil, err
		}
		xv := float64(t.Unix())
		ticks[i] = chart.Tick{Value: xv}
		if (i+1)%5 == 0 || i == 0 {
			ticks[i].Label = fmt.Sprintf("%v %s", t.Day(), months[t.Month()-1])
		}
		c.XValues[i] = xv
		d.XValues[i] = xv
		r.XValues[i] = xv
		h.XValues[i] = xv

		c.YValues[i] = float64(v.Confirmed)
		d.YValues[i] = float64(v.Deaths)
		r.YValues[i] = float64(v.Recovered)
		h.YValues[i] = float64(v.Hospitalized)

		if v.Confirmed > max {
			max = v.Confirmed
		}
	}
	graph := chart.Chart{
		Font:   f,
		Height: 300,
		Width:  600,
		XAxis: chart.XAxis{
			Ticks: ticks,
			Style: chart.StyleShow(),
		},
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: float64(max),
			},
			Style: chart.StyleShow(),
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf("%s", humanize.Comma(int64(vf)))
				}
				return ""
			},
		},
		Series: []chart.Series{c, d, r, h},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	buf := new(bytes.Buffer)

	err = graph.Render(chart.PNG, buf)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return buf, nil
}
