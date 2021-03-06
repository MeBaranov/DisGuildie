package message

import (
	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"
)

type Message interface {
	GuildId() string
	ChannelId() string
	Mentions() []string
	AuthorId() string
	Author() (*database.User, error)
	Money() (*database.Money, error)
	AuthorPermissions() (int, error)
	FullMessage() string

	GuildMembers() (map[string]string, error)
	GuildMembersWithRole(string) (map[string]string, error)
	UserRoles(string) ([]string, error)
	GetRoleId(string) (string, error)

	CurSegment() string
	PeekSegment() string
	LeftOverSegments() string
	MoreSegments() bool

	SendMessage(string, ...interface{})

	CheckGuildModificationPermissions(uuid.UUID) (bool, error)
	CheckUserModificationPermissions(uid string) (bool, error)
}
