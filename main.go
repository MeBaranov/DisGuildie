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
		price     int
		duration  int
	)

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&superUser, "su", "", "Super User")
	flag.IntVar(&price, "p", 1000, "Price (Cents)")
	flag.IntVar(&duration, "d", 30, "Free trial duration (days)")
	flag.Parse()

	if token == "" {
		token = os.Getenv("BOT_TOKEN")

		if token == "" {
			fmt.Println("No token provided. Please run this app with \"-t <bot token>\" or with BOT_TOKEN environment variable set")
			return
		}
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

	p, err := processor.New(dataProvider, token, discordgo.MakeIntent(intent), timeout, &superUser, price, h)
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
