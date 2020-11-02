package helpers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/utility"
)

type AdminUserProcessor struct {
	prov  database.DataProvider
	funcs map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate)
}

func NewAdminUserProcessor(prov database.DataProvider) MessageProcessor {
	ap := &AdminUserProcessor{prov: prov}
	ap.funcs = map[string]func(*discordgo.Session, *string, *discordgo.MessageCreate){
		"h":        ap.help,
		"help":     ap.help,
		"r":        ap.register,
		"register": ap.register,
	}
	return ap
}

func (ap *AdminUserProcessor) ProcessMessage(s *discordgo.Session, m *string, mc *discordgo.MessageCreate) {
	cmd, obj := utility.NextCommand(m)

	f, ok := ap.funcs[cmd]
	if !ok {
		rv := fmt.Sprintf("Unknown user administration command \"%v\". Send \"!g admin user help\" or \"!g a u h\" for help", mc.Message.Content)
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	f(s, &obj, mc)
}

func (ap *AdminUserProcessor) register(s *discordgo.Session, m *string, mc *discordgo.MessageCreate) {
	rv := "Oh my, what do I do now?\n"
	rv += "Somehow parse this: ```" + *m + "```"
	fmt.Println("Message: ", *m)
	rv += "With this: " + fmt.Sprintln(mc.Mentions)
	if len(mc.Mentions) > 0 {
		rv += "\t" + fmt.Sprintln(mc.Mentions[0])
	}

	go utility.SendMonitored(s, &mc.ChannelID, &rv)
}

func (ap *AdminUserProcessor) help(s *discordgo.Session, _ *string, mc *discordgo.MessageCreate) {
	rv := "Here's a list of user management commands you're allowed to do:\n"

	perm, err := utility.GetPermissions(s, mc, ap.prov)
	if err != nil {
		rv += "Some error happened while getting permissions: " + err.Error()
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	if perm&database.CharsPermissions == 0 {
		rv += "Sorry, none. Ask leaders to let you do more"
		go utility.SendMonitored(s, &mc.ChannelID, &rv)
		return
	}

	rv += "\t -- \"!g admin user register <mention user>\" (\"!g a u r <user>\") - Register user in the system\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user register all\" (\"!g a u r all\") - Register all users from guild in the system\n"
	}
	rv += "\t -- \"!g admin user remove <mention user>\" (\"!g a u remove <user>\") - Remove user from the system\n"
	rv += "\t -- \"!g admin user remove <Discord ID>\" (\"!g a u remove <Discord ID>\") - Remove user from the system by discord ID\n"
	if perm&database.EditGuildCharsPerm != 0 {
		rv += "\t -- \"!g admin user sync\" (\"!g a u s\") - Register all users in guild in the system. And remove all users no longer in guild\n"
	}
	rv += "\t -- \"!g admin user assign <mention user> <sub-guild name>\" (\"!g a u a <user> <name>\") - Move user to a sub-guild\n"
	rv += "\t -- \"!g admin user permissions <mention user>\" (\"!g a u p <user>\") - Re-synchronize user roles and permissions\n"

	go utility.SendMonitored(s, &mc.ChannelID, &rv)
}
