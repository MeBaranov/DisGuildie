package processor

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/processor/helpers"
)

type Processor struct {
	provider database.DataProvider
	funcs    map[string]func(message.Message)
	s        *discordgo.Session
	rc       chan bool
}

func New(prov database.DataProvider, token string, intent *discordgo.Intent, timeout time.Duration) (*Processor, error) {
	admin := helpers.NewAdminProcessor(prov)
	proc := &Processor{provider: prov, rc: make(chan bool)}

	proc.funcs = map[string]func(message.Message){
		"help":  proc.help,
		"h":     proc.help,
		"admin": admin.ProcessMessage,
		"a":     admin.ProcessMessage,
	}

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	proc.s = s
	s.Identify.Intents = intent

	fmt.Println("Bot created successfully")
	s.AddHandler(proc.ready)
	s.AddHandler(proc.messageCreate)
	// TODO: Add GuildCreate handler

	err = s.Open()
	if err != nil {
		return nil, err
	}

	fmt.Println("Bot connection opened")

	select {
	case res := <-proc.rc:
		if !res {
			return nil, errors.New("Session did not get ready.")
		}
		fmt.Println("Connected successfully")
	case <-time.After(timeout):
		return nil, errors.New("Session did not start properly.")
	}

	return proc, nil
}

func (proc *Processor) Close() {
	proc.s.Close()
}

func (proc *Processor) processMessage(m message.Message) {
	cmd := m.CurSegment()

	f, ok := proc.funcs[cmd]
	if !ok {
		rv := fmt.Sprintf("Unknown command \"%v\". Send \"!g help\" or \"!g h\" for help", m.FullMessage())
		go m.SendMessage(&rv)
		return
	}

	f(m)
}

func (proc *Processor) help(m message.Message) {
	rv := "Here is a list of commands you are allowed to use:\n"

	p, err := m.AuthorPermissions()
	if err != nil {
		rv += "Could not get your permissions. Error:\n" + err.Error()
		go m.SendMessage(&rv)
		return
	}

	if p > 0 {
		rv += "\t-- \"!g admin\" (\"!g a\") - administrative actions"
	}

	go m.SendMessage(&rv)
}

func (proc *Processor) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!g") || m.Author.Bot {
		return
	}

	msg := message.New(s, m, proc.provider)
	msg.CurSegment()

	proc.processMessage(msg)
}

func (proc *Processor) ready(s *discordgo.Session, r *discordgo.Ready) {
	for _, g := range r.Guilds {
		if _, err := proc.provider.GetGuildD(g.ID); err != nil {
			if err.Code != database.GuildNotFound {
				fmt.Println("Something has gone wrong while getting guilds. Error:", err.Error())
				proc.rc <- false
				return
			}

			dbg := &database.Guild{
				DiscordId: g.ID,
				Name:      g.Name,
				Stats:     make(map[string]string),
			}
			_, err = proc.provider.AddGuild(dbg)
			if err != nil {
				fmt.Println("Something has gone wrong while adding guild", g.Name, "with ID", g.ID, ". Error:", err.Error())
				proc.rc <- false
				return
			}
		}
	}
	proc.rc <- true
}
