package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/database/memory"

	"github.com/bwmarrin/discordgo"

	"github.com/mebaranov/disguildie/processor"
	"github.com/mebaranov/disguildie/processor/helpers"
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
var messageProcessor helpers.MessageProcessor
var dataProvider database.DataProvider

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
	var superUser string
	flag.StringVar(&superUser, "su", "", "Super User")
	flag.Parse()

	if token == "" {
		token = os.Getenv("BOT_TOKEN")

		if token == "" {
			err = errors.New("No token provided. Please run this app with \"-t <bot token>\" or with BOT_TOKEN environment variable set")
		}
	}
	utility.SuperUserID = superUser

	dataProvider = memory.NewMemoryDb()
	messageProcessor = processor.New(dataProvider)

	return
}

func ready(s *discordgo.Session, r *discordgo.Ready) {
	readyChannel <- (r != nil)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!g") || m.Author.Bot {
		return
	}

	_, msg := utility.NextCommand(&(m.Content))
	messageProcessor.ProcessMessage(s, &msg, m)
}
