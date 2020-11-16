package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mebaranov/disguildie/database/memory"
	"github.com/mebaranov/disguildie/processor"

	"github.com/bwmarrin/discordgo"
)

const timeout = time.Second * 10

var intents = [...]discordgo.Intent{
	discordgo.IntentsGuilds,
	discordgo.IntentsGuildMembers,
	discordgo.IntentsGuildMessageReactions,
	discordgo.IntentsGuildMessages,
	discordgo.IntentsGuildPresences,
	discordgo.IntentsGuilds,
}

func main() {
	var (
		token     string
		superUser string
		owner     string
		price     int
		duration  int
		link      string
	)

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&superUser, "su", "", "Super User")
	flag.StringVar(&owner, "o", "", "Owner discord")
	flag.IntVar(&price, "p", -1, "Price (Cents)")
	flag.IntVar(&duration, "d", -1, "Free trial duration (days)")
	flag.StringVar(&link, "l", "", "Payment link template")
	flag.Parse()

	if token == "" {
		token = os.Getenv("BOT_TOKEN")
	}
	if superUser == "" {
		superUser = os.Getenv("BOT_SUPER_USER")
	}
	if owner == "" {
		owner = os.Getenv("BOT_OWNER")
	}
	if link == "" {
		link = os.Getenv("BOT_PAYMENT_LINK")
	}
	if price == -1 {
		tmp := os.Getenv("BOT_PRICE")
		var err error
		if tmp != "" {
			if price, err = strconv.Atoi(tmp); err != nil {
				fmt.Println("Could not parse price:", tmp, ". Error:", err.Error())
			}
		} else {
			price = 1000
		}
	}
	if duration == -1 {
		tmp := os.Getenv("BOT_FREE_DURATION")
		var err error
		if tmp != "" {
			if duration, err = strconv.Atoi(tmp); err != nil {
				fmt.Println("Could not parse duration:", tmp, ". Error:", err.Error())
			}
		} else {
			duration = 30
		}
	}

	if token == "" {
		fmt.Println("No token provided. Please run this app with \"-t <bot token>\" or with BOT_TOKEN environment variable set")
		return
	}

	dataProvider := memory.NewMemoryDb()
	intent := discordgo.IntentsNone
	for _, i := range intents {
		intent |= i
	}

	hoursStr := strconv.Itoa(duration * 24)
	h, err := time.ParseDuration(hoursStr + "h")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	p, err := processor.New(dataProvider, token, discordgo.MakeIntent(intent), timeout, &superUser, price, h, owner, link)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer p.Close()

	runningChannel := make(chan bool)
	// go cmd.Start(runningChannel)
	<-runningChannel

	fmt.Println("Exitting. Hoping to do it gracefully.")
}
