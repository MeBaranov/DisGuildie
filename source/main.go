package main

import (
	"flag"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var ch chan bool
var token string

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	if token == "" {
		fmt.Println("No token provided. Please run: this app with \"-t <bot token>\"")
		return
	}

	d, err := discordgo.New("Bot " + token)

	if err != nil {
		fmt.Println("Could not create bot. Error:", err)
	}

	ch = make(chan bool)

	fmt.Println("Bot created successfully")
	d.AddHandler(ready)

	d.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuilds)
	err = d.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	if _, ok := <-ch; !ok {
		fmt.Println("Could not read from chan")
	}

	fmt.Println("What else do you want from me?")
	d.Close()
}

func ready(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Println("Guilds:")
	for _, g := range r.Guilds {
		fmt.Println(g)
	}
	s.Close()
	ch <- true
}
