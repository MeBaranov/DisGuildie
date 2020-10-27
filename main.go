package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/mebaranov/disguildie/processor"
	"github.com/mebaranov/disguildie/utility"
)

const readyTimeout = time.Second * 10

var intents = [...]discordgo.Intent{
	discordgo.IntentsGuilds,
	discordgo.IntentsGuildMembers,
	discordgo.IntentsGuildMessageReactions,
	discordgo.IntentsGuildMessages,
	discordgo.IntentsGuildPresences,
	discordgo.IntentsGuilds,
}

var readyChannel chan bool = make(chan bool)
var messageProcessor processor.MessageProcessor

func main() {
	token, err := setUp()

	if err != nil {
		fmt.Println("Could not start. Error:", err)
		return
	}

	d, err := discordgo.New("Bot " + token)

	if err != nil {
		fmt.Println("Could not create bot. Error:", err)
		return
	}

	fmt.Println("Bot created successfully")
	d.AddHandler(ready)
	d.AddHandler(messageCreate)

	intent := discordgo.IntentsNone
	for _, i := range intents {
		intent |= i
	}

	d.Identify.Intents = discordgo.MakeIntent(intent)

	err = d.Open()
	if err != nil {
		fmt.Println("Could not open session. Error: ", err)
		return
	}

	defer d.Close()
	fmt.Println("Bot connection opened")

	select {
	case res := <-readyChannel:
		if !res {
			fmt.Println("Session did not get ready.")
			return
		}
		fmt.Println("Connected successfully")
	case <-time.After(readyTimeout):
		fmt.Println("Session did not start properly. ")
		return
	}

	runningChannel := make(chan bool)
	// go cmd.Start(runningChannel)
	<-runningChannel

	fmt.Println("Exitting. Hoping to do it gracefully.")
}

func setUp() (token string, err error) {
	err = nil
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()

	if token == "" {
		token = os.Getenv("BOT_TOKEN")

		if token == "" {
			err = errors.New("No token provided. Please run this app with \"-t <bot token>\" or with BOT_TOKEN environment variable set")
		}
	}

	messageProcessor = &processor.Processor{}

	return
}

func ready(s *discordgo.Session, r *discordgo.Ready) {
	readyChannel <- (r != nil)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!g") {
		return
	}

	fmt.Println("Content:", m.Content)
	fmt.Println("Message:", m.Mentions)
	_, msg := utility.NextCommand(&(m.Content))
	messageProcessor.ProcessMessage(s, &(m.ChannelID), &msg, m)
}
