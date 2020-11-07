package message

type Message interface {
	GuildId() string
	ChannelId() string
	Mentions() []string
	Author() string
	AuthorPermissions() (int, error)
	FullMessage() string

	GuildMembers() (map[string]string, error)

	CurSegment() string
	MoreSegments() bool

	SendMessage(*string)
}
