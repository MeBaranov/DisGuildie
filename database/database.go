package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Void struct{}

var Member Void

type Guild struct {
	GuildId          uuid.UUID
	ParentId         uuid.UUID
	TopLevelParentId uuid.UUID
	DiscordId        string
	Name             string
	Stats            map[string]string
	ChildNames       map[string]Void
}

type User struct {
	UserId    uuid.UUID
	DiscordId string
	Guilds    map[uuid.UUID]int
}

type Character struct {
	CharId string
	UserId uuid.UUID
	Name   string
	Main   bool
	Body   map[string]interface{}
}

type Role struct {
	GuildId     string
	Id          string
	Permissions int
}

type Money struct {
	GuildId string
	UserId  string
	ValidTo time.Time
	Price   int
}

type DataProvider interface {
	AddGuild(g *Guild) (*Guild, error)
	GetGuild(g uuid.UUID) (*Guild, error)
	GetGuildD(d string) (*Guild, error)
	RenameGuild(g uuid.UUID, name string) (*Guild, error)
	AddGuildStat(g uuid.UUID, n string, t string) (*Guild, error)
	RemoveGuildStat(g uuid.UUID, n string) (*Guild, error)
	RemoveGuild(g uuid.UUID) (*Guild, error)
	RemoveGuildD(d string) (*Guild, error)

	AddUser(u *User, g uuid.UUID, p int) (*User, error)
	GetUserD(d string) (*User, error)
	SetUserPermissions(u string, g uuid.UUID, p int) (*User, error)
	RemoveUserD(d string, g uuid.UUID) (*User, error)
	EraseUserD(d string) (*User, error)

	AddCharacter(c *Character) (*Character, error)
	GetCharacters(u uuid.UUID) ([]*Character, error)
	GetMainCharacter(u uuid.UUID) (*Character, error)
	GetCharacter(u uuid.UUID, n string) (*Character, error)
	RenameCharacter(u uuid.UUID, old string, name string) (*Character, error)
	ChangeMainCharacter(u uuid.UUID, name string) (*Character, error)
	SetCharacterStat(u uuid.UUID, name string, s string, v interface{}) (*Character, error)
	ChangeCharacterOwner(old uuid.UUID, name string, u uuid.UUID) (*Character, error)
	RemoveCharacterStat(u uuid.UUID, name string, s string) (*Character, error)
	RemoveCharacter(u uuid.UUID, name string) (*Character, error)

	AddRole(r *Role) (*Role, error)
	GetRole(g string, r string) (*Role, error)
	GetGuildRoles(g string) ([]*Role, error)
	SetRolePermissions(g string, r string, p int) (*Role, error)
	RemoveRole(g string, r string) (*Role, error)

	AddMoney(m *Money) (*Money, error)
	GetMoney(g string) (*Money, error)
	ChangeMoneyOwner(g string, u string) (*Money, error)
	SetMoneyValid(g string, t time.Time) (*Money, error)

	Export() ([]byte, error)
	Import(b []byte) error
}

type ErrorCode int
type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error '%v': %v", e.Code, e.Message)
}

const (
	_ = iota
	ExternalError
	ConnectionErroruser
	InvalidGuildDefinition
	InvalidDatabaseState
	GuildAlreadyRegistered
	SubguildNameTaken
	GuildNotFound
	GuildLevelError
	StatNameConflict
	StatNotFound
	UserNotFound
	UserNotInGuild
	UserAlreadyInGuild
	WrongUserInput
	NoMainCharacterSpecified
	CharacterNotFound
	CharacterNameTaken
	UserHasCharacter
	RoleAlreadyExists
	RoleNotFound
	MoneyAlreadyRegistered
	MoneyNotFound
)
