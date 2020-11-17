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
	"github.com/mebaranov/disguildie/processor/helpers/admin"
	"github.com/mebaranov/disguildie/processor/helpers/user"
)

type Processor struct {
	helpers.BaseMessageProcessor
	s            *discordgo.Session
	rc           chan bool
	superUser    *string
	price        int
	freeDuration time.Duration
	ownerDiscord string
	paymentLink  string
}

func New(
	prov database.DataProvider,
	token string,
	intent *discordgo.Intent,
	timeout time.Duration,
	superUser *string,
	price int,
	freeDuration time.Duration,
	ownerDiscord string,
	paymentLink string) (*Processor, error) {

	admin := admin.NewAdminProcessor(prov)
	char := user.NewCharProcessor(prov)
	list := user.NewListProcessor(prov)
	owner := user.NewOwnerProcessor(prov)
	stats := user.NewStatsProcessor(prov)
	top := user.NewTopProcessor(prov)
	hierarchy := user.NewHierarchyProcessor(prov)
	gdpr := user.NewGdprProcessor(prov)

	proc := &Processor{
		rc:           make(chan bool),
		superUser:    superUser,
		price:        price,
		freeDuration: freeDuration,
		ownerDiscord: ownerDiscord,
		paymentLink:  paymentLink,
	}

	proc.Prov = prov
	proc.Funcs = map[string]func(message.Message) (string, error){
		"help":      proc.help,
		"h":         proc.help,
		"admin":     admin.ProcessMessage,
		"a":         admin.ProcessMessage,
		"list":      list.ProcessMessage,
		"l":         list.ProcessMessage,
		"stat":      stats.ProcessMessage,
		"s":         stats.ProcessMessage,
		"char":      char.ProcessMessage,
		"c":         char.ProcessMessage,
		"top":       top.ProcessMessage,
		"t":         top.ProcessMessage,
		"owner":     owner.ProcessMessage,
		"o":         owner.ProcessMessage,
		"hierarchy": hierarchy.ProcessMessage,
		"hi":        hierarchy.ProcessMessage,
		"gdpr":      gdpr.ProcessMessage,
		"g":         gdpr.ProcessMessage,
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

func (proc *Processor) help(m message.Message) (string, error) {
	rv := "Here is a list of commands you are allowed to use:\n"

	p, err := m.AuthorPermissions()
	if err != nil {
		return "getting author permissions", err
	}

	if p > 0 {
		rv += "\t-- \"!g admin\" (\"!g a\") - administrative\n"
	}
	rv += "\t-- \"!g char\" (\"!g c\") - character management\n"
	rv += "\t-- \"!g list\" (\"!g l\") - list characters\n"
	rv += "\t-- \"!g owner\" (\"!g o\") - get owner(s) of character\n"
	rv += "\t-- \"!g stat\" (\"!g s\") - stats management\n"
	rv += "\t-- \"!g top\" (\"!g t\") - guild tops\n"
	rv += "\t-- \"!g hierarchy\" (\"!g hi\") - sub-guilds structure\n"
	rv += "\t-- \"!g gdpr\" - GDPR-related\n"

	rv += "\nThis bot is distributed under Apache2 license. You can find source code on github: https://github.com/MeBaranov/DisGuildie\n"
	rv += "To contact the owner you can use github link above"
	if proc.ownerDiscord != "" {
		rv += ", or discord: " + proc.ownerDiscord
	}
	rv += "\n"

	mon, err := m.Money()
	if err != nil {
		return "getting payments", err
	}

	if mon.Price == 0 {
		rv += "You're using this bot for free. Congratulations!"
		return rv, nil
	} else if mon.ValidTo.After(time.Now()) {
		diff := mon.ValidTo.Sub(time.Now()).Hours()
		diffi := int(diff / 24)
		rv += fmt.Sprintf("Your bot is payed for and will be active for %v days.\n", diffi)
	} else {
		diff := time.Now().Sub(mon.ValidTo).Hours()
		diffi := int(diff / 24)
		rv += fmt.Sprintf("Your subscription has ended %v days ago.\n", diffi)
	}

	if proc.paymentLink != "" {
		rv += "You can extend your subscription using the following link:\n" + fmt.Sprintf(proc.paymentLink, mon.GuildId)
	}

	return rv, nil
}

func (proc *Processor) ready(s *discordgo.Session, r *discordgo.Ready) {
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

	msg := message.New(s, m, proc.Prov, proc.superUser)

	msg.CurSegment()

	mon, err := msg.Money()
	if err != nil {
		msg.SendMessage("Could not validate guild: %v", err.Error())
		return
	}
	if time.Now().After(mon.ValidTo) && mon.Price > 0 && msg.PeekSegment() != "h" && msg.PeekSegment() != "help" {
		msg.SendMessage("Guild subscription is out of date. Please, extend subscription. Use \"!g h\" for more details", mon.UserId)
		return
	}

	if msg.AuthorId() != *(proc.superUser) {
		_, err = msg.Author()
		if err != nil {
			msg.SendMessage("Could not validate your registration: %v", err.Error())
			return
		}
	}

	rv, err := proc.ProcessMessage(msg)
	if err != nil {
		msg.SendMessage("Error %v: %v", rv, err)
		return
	}
	msg.SendMessage(rv)
}

func (proc *Processor) tryRegisterGuild(g *discordgo.Guild) error {
	if _, err := proc.Prov.GetGuildD(g.ID); err != nil {
		dbErr := database.ErrToDbErr(err)
		if dbErr == nil || dbErr.Code != database.GuildNotFound {
			return errors.New("Error while getting guilds: " + err.Error())
		}

		dbg := &database.Guild{
			DiscordId: g.ID,
			Name:      "main",
		}

		_, err = proc.Prov.AddGuild(dbg)
		if err != nil {
			return errors.New(fmt.Sprintln("Error while adding guild", g.Name, "with ID", g.ID, ":", err.Error()))
		}

		if _, err := proc.Prov.GetMoney(g.ID); err != nil {
			dbErr := database.ErrToDbErr(err)
			if dbErr == nil || dbErr.Code != database.MoneyNotFound {
				return errors.New("Something has gone wrong while getting guilds. Error: " + err.Error())
			}

			money := &database.Money{
				GuildId: g.ID,
				Price:   proc.price,
				ValidTo: time.Now().Add(proc.freeDuration),
				UserId:  g.OwnerID,
			}
			_, err = proc.Prov.AddMoney(money)
			if err != nil {
				return errors.New(fmt.Sprintln("Something has gone wrong while adding money", g.Name, "with ID", g.ID, ". Error:", err.Error()))
			}
		}
	}

	return nil
}
