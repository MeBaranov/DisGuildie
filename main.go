package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/mebaranov/disguildie/database/memory"
	"github.com/mebaranov/disguildie/processor"

	"github.com/bwmarrin/discordgo"

	"github.com/mebaranov/disguildie/utility"
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
	var token string
	var superUser string

	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&superUser, "su", "", "Super User")
	flag.Parse()

	if token == "" {
		token = os.Getenv("BOT_TOKEN")

		if token == "" {
			fmt.Println("No token provided. Please run this app with \"-t <bot token>\" or with BOT_TOKEN environment variable set")
			return
		}
	}
	utility.SuperUserID = superUser

	dataProvider := memory.NewMemoryDb()
	intent := discordgo.IntentsNone
	for _, i := range intents {
		intent |= i
	}

	p, err := processor.New(dataProvider, token, discordgo.MakeIntent(intent), timeout)
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
