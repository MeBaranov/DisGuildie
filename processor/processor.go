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
	provider     database.DataProvider
	funcs        map[string]func(message.Message)
	s            *discordgo.Session
	rc           chan bool
	superUser    *string
	price        int
	freeDuration time.Duration
}

func New(
	prov database.DataProvider,
	token string,
	intent *discordgo.Intent,
	timeout time.Duration,
	superUser *string,
	price int,
	freeDuration time.Duration) (*Processor, error) {

	admin := helpers.NewAdminProcessor(prov)
	proc := &Processor{
		provider:     prov,
		rc:           make(chan bool),
		superUser:    superUser,
		price:        price,
		freeDuration: freeDuration,
	}

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
	s.AddHandler(proc.guildCreate)

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
		m.SendMessage("Unknown command \"%v\". Send \"!g help\" or \"!g h\" for help", m.FullMessage())
		return
	}

	f(m)
}

func (proc *Processor) help(m message.Message) {
	rv := "Here is a list of commands you are allowed to use:\n"

	p, err := m.AuthorPermissions()
	if err != nil {
		rv += "Could not get your permissions. Error:\n" + err.Error()
		m.SendMessage(rv)
		return
	}

	if p > 0 {
		rv += "\t-- \"!g admin\" (\"!g a\") - administrative actions"
	}

	m.SendMessage(rv)
}

func (proc *Processor) ready(s *discordgo.Session, r *discordgo.Ready) {
	for _, g := range r.Guilds {
		if err := proc.tryRegisterGuild(g); err != nil {
			fmt.Print(err.Error())
			proc.rc <- false
			return
		}
	}
	proc.rc <- true
}

func (proc *Processor) guildCreate(s *discordgo.Session, r *discordgo.GuildCreate) {
	if err := proc.tryRegisterGuild(r.Guild); err != nil {
		fmt.Printf("Critical: Could not add guild with ID: '%v', Name: '%v'", r.Guild.ID, r.Guild.Name)
		return
	}
	fmt.Printf("Added guild with ID: '%v', Name: '%v'", r.Guild.ID, r.Guild.Name)
}

func (proc *Processor) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!g") || m.Author.Bot {
		return
	}

	msg := message.New(s, m, proc.provider, proc.superUser)

	mon, err := msg.Money()
	if err != nil {
		msg.SendMessage("Could not validate guild: %v", err.Error())
		return
	}
	if time.Now().After(mon.ValidTo) {
		msg.SendMessage("Guild subscription is out of date. Please, ask <@!%v> to extend subscription.", mon.UserId)
		return
	}

	if msg.AuthorId() != *(proc.superUser) {
		_, err = msg.Author()
		if err != nil {
			msg.SendMessage("Could not validate your registration: %v", err.Error())
			return
		}
	}

	msg.CurSegment()

	proc.processMessage(msg)
}

func (proc *Processor) tryRegisterGuild(g *discordgo.Guild) error {
	if _, err := proc.provider.GetGuildD(g.ID); err != nil {
		if err.Code != database.GuildNotFound {
			return errors.New("Error while getting guilds: " + err.Error())
		}

		dbg := &database.Guild{
			DiscordId: g.ID,
			Name:      g.Name,
			Stats:     make(map[string]string),
		}

		_, err = proc.provider.AddGuild(dbg)
		if err != nil {
			return errors.New(fmt.Sprintln("Error while adding guild", g.Name, "with ID", g.ID, ":", err.Error()))
		}

		if _, err := proc.provider.GetMoney(g.ID); err != nil {
			if err.Code != database.MoneyNotFound {
				return errors.New("Something has gone wrong while getting guilds. Error: " + err.Error())
			}

			money := &database.Money{
				GuildId: g.ID,
				Price:   proc.price,
				ValidTo: time.Now().Add(proc.freeDuration),
				UserId:  g.OwnerID,
			}
			_, err = proc.provider.AddMoney(money)
			if err != nil {
				return errors.New(fmt.Sprintln("Something has gone wrong while adding money", g.Name, "with ID", g.ID, ". Error:", err.Error()))
			}
		}
	}

	return nil
}
