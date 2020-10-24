package helpers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/mebaranov/disguildie/utility"
)

var help = `There's a list of stuff I can do. But first you'll need some notation explanation.
This message is really long, so it will be split into several messages.
Anything in square brackets '[]' should be replaced with relevant values. [Username] means 'put any user name here'
Things in round brackets '()' should be used as is. (TypeMe) means 'put TypeMe here'
Anything separated by pipe '|' requests to select one of options. (gs|name|stuff) means 'put one of those three here'
Pipe at the end of grouping means that whole group can be ignord. (gs|name|) means 'put gs, or name, or nothing'
Some commands are market as **[Officer]**. They require officer role to work.
Let's begin.
1) Help
	-) !help - will show this help
2) Settings
2) Accounts
    -) !leaderboard - will show leaderboard by Gearscore
    -) !loserboard - will show loserboard by Gearscore
    -) !mystats - will show known info about you
    -) !stats [Username] - will show info about different user
    -) !change (level|class|ign|gear|gs) {Value} - will change your parameter accordingly
        --) level and class are obvious
        --) ign is in game name
        --) gear is a link to screenshot with your character info window
        --) gs is your gearscore
    -) **[Officer]** !duespaid {username} (true|false) - Sets due of user to paid or not
    -) **[Officer]** !duespaid {all|none} - Sets dues of all users in guild to paid or not
    -) **[Officer]** !duesreminder - Reminds all those who haven't paid dues to pay them
    -) **[Officer]** !whohasntpaid - Lists those who haven't paid dues
    -) **[Officer]** !roster (name|gs|guild|dkp|),(asc|desc|),{top} - Shows all or top N of guild members, ordered by specified field, ascending or descending
    -) **[Officer]** !change (guild) {Value} - Changes officer's own guild to a different one
    -) **[Officer]** !change {User} (level|class|ign|gear|gs|guild|role) {Value} - Change specified users data. Used with @ discord reference
    -) **[Officer]** !changebyign {User} (level|class|ign|gear|gs|guild|role) {Value} - Change specified users data. Used with in game name
    -) **[Officer]** !adddkp {User} {Amount} - Add dkp
    -) **[Officer]** !adduser {User} - Add user to the system
    -) **[Officer]** !removeuser {User} - Remove user from the system. Used with @ discord reference
    -) **[Officer]** !removeuserbyign {User} - Remove user from the system. Used with in game name
    -) **[Officer]** !cleanupusers - Removes stale users from the system. They are in the system but are not in discord
    -) **[Officer]** **[Leader only]** !changeallguild {Value} - Move all users to a new guild
    -) **[Officer]** **[Leader only]** !addallusers - Add all users from current discord server to the system
	-) **[Officer]** **[Leader only]** !clearallusers - Removes all users from the system
3) Administration
	-) 
That's all for now. Good luck and have fun :)`

type HelpProcessor struct{}

func (hp *HelpProcessor) ProcessMessage(s *discordgo.Session, c *string, m *string, mc *discordgo.MessageCreate) {
	utility.SendMonitored(s, c, &help)
}
