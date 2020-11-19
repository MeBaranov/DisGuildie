package tests

import (
	"github.com/google/uuid"
	"github.com/mebaranov/disguildie/database"
	"github.com/mebaranov/disguildie/message"
	"github.com/mebaranov/disguildie/utility"
)

type TestMessage struct {
	GuildIdMock           func() string
	ChannelIdMock         func() string
	MentionsMock          func() []string
	AuthorIdMock          func() string
	AuthorMock            func() (*database.User, error)
	MoneyMock             func() (*database.Money, error)
	AuthorPermissionsMock func() (int, error)
	FullMessageMock       func() string

	GuildMembersMock         func() (map[string]string, error)
	GuildMembersWithRoleMock func(string) (map[string]string, error)
	UserRolesMock            func(string) ([]string, error)
	GetRoleIdMock            func(string) (string, error)

	CurSegmentMock       func() string
	PeekSegmentMock      func() string
	LeftOverSegmentsMock func() string
	MoreSegmentsMock     func() bool

	SendMessageMock func(string, ...interface{})

	CheckGuildModificationPermissionsMock func(gid uuid.UUID) (bool, error)
	CheckUserModificationPermissionsMock  func(uid string) (bool, error)

	CurMsg string
}

func New() message.Message {
	return &TestMessage{}
}

func validator() {
	var _ message.Message = &TestMessage{}
}

func (tm *TestMessage) GuildId() string {
	return tm.GuildIdMock()
}

func (tm *TestMessage) AuthorId() string {
	return tm.AuthorIdMock()
}

func (tm *TestMessage) ChannelId() string {
	return tm.ChannelIdMock()
}

func (tm *TestMessage) FullMessage() string {
	return tm.FullMessageMock()
}

func (tm *TestMessage) Author() (*database.User, error) {
	return tm.AuthorMock()
}

func (tm *TestMessage) AuthorPermissions() (int, error) {
	return tm.AuthorPermissionsMock()
}

func (tm *TestMessage) Money() (*database.Money, error) {
	return tm.MoneyMock()
}

func (tm *TestMessage) Mentions() []string {
	return tm.MentionsMock()
}

func (tm *TestMessage) GuildMembers() (map[string]string, error) {
	return tm.GuildMembersMock()
}

func (tm *TestMessage) GuildMembersWithRole(r string) (map[string]string, error) {
	return tm.GuildMembersWithRoleMock(r)
}

func (tm *TestMessage) CurSegment() string {
	var rv string
	for rv == "" && tm.CurMsg != "" {
		rv, tm.CurMsg = utility.NextCommand(&tm.CurMsg)
	}

	return rv
}

func (tm *TestMessage) PeekSegment() string {
	var rv string
	tmp := tm.CurMsg
	for rv == "" && tmp != "" {
		rv, tmp = utility.NextCommand(&tmp)
	}

	return rv
}

func (tm *TestMessage) LeftOverSegments() string {
	return tm.CurMsg
}

func (tm *TestMessage) MoreSegments() bool {
	return tm.CurMsg == ""
}

func (tm *TestMessage) SendMessage(s string, strs ...interface{}) {
	tm.SendMessageMock(s, strs)
}

func (tm *TestMessage) UserRoles(id string) ([]string, error) {
	return tm.UserRolesMock(id)
}

func (tm *TestMessage) GetRoleId(name string) (string, error) {
	return tm.GetRoleIdMock(name)
}

func (tm *TestMessage) CheckGuildModificationPermissions(gid uuid.UUID) (bool, error) {
	return tm.CheckGuildModificationPermissionsMock(gid)
}

func (tm *TestMessage) CheckUserModificationPermissions(uid string) (bool, error) {
	return tm.CheckUserModificationPermissionsMock(uid)
}
