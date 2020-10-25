package database

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Void struct{}

var Member Void

type Guild struct {
	GuildId    uuid.UUID
	ParentId   uuid.UUID
	DiscordId  string
	Name       string
	Stats      map[string]string
	ChildNames map[string]Void
}

type User struct {
	UserId      uuid.UUID
	DiscordId   string
	GuildIds    map[uuid.UUID]Void
	Permissions int
}

type Character struct {
	CharId uuid.UUID
	UserId uuid.UUID
	Name   string
	IsMain bool
	Body   map[string]interface{}
}

type Role struct {
	GuildId     uuid.UUID
	Users       []uuid.UUID
	DiscordId   string
	Permissions int
}

type Money struct {
	GuildId uuid.UUID
	UserId  uuid.UUID
	ValidTo time.Time
	Price   int
}

type DataProvider interface {
	AddGuild(g Guild) *Guild
	GetGuild(d string) *Guild
	GetGuildD(g uuid.UUID) *Guild
	RenameGuild(g uuid.UUID, name string) *Guild
	RestatGuild(g uuid.UUID, name string) *Guild
	RemoveGuild(g uuid.UUID) *Guild
	RemoveGuildD(d string) *Guild

	AddUser(u User) *User
	GetUser(d string, g uuid.UUID) *User
	SetUserPermissions(u uuid.UUID, p int)
	RemoveUser(u uuid.UUID) *User
	RemoveUserD(d string) *User

	AddCharacter(u uuid.UUID, name string, IsMain bool)
	GetCharacters(u uuid.UUID) *[]Character
	GetMainCharacter(u uuid.UUID) *Character
	GetCharacter(c uuid.UUID) *Character
	RenameCharacter(c uuid.UUID, name string) *Character
	ChangeMainCharacter(c uuid.UUID, IsMain bool) *Character
	SetCharacterStat(c uuid.UUID, s string, v interface{}) *Character
	ChangeCharacterOwner(c uuid.UUID, u uuid.UUID) *Character
	RemoveCharacterStat(c uuid.UUID, s string) *Character
	RemoveCharacter(c uuid.UUID) *Character

	AddRole(r Role) *Role
	GetRole(g uuid.UUID, d string) *Role
	GetRolesUser(u uuid.UUID) *[]Role
	GetRolesGuild(g uuid.UUID) *[]Role
	SetRolePermissions(r uuid.UUID) *Role
	RemoveRole(r uuid.UUID) *Role

	AddMoney(m Money) Money
	GetMoneyGuid(g uuid.UUID) Money
	ChangeGuildOwner(g uuid.UUID, u uuid.UUID) Money
	SetMoneyValid(m uuid.UUID, t time.Time) Money

	Export() string
	Import(s string)
}

type DbError struct {
	Code    int
	Message string
}

func (e *DbError) Error() string {
	return fmt.Sprintf("Error '%v': %v", e.Code, e.Message)
}

const (
	_ = iota
	InvalidGuildDefinition
	InvalidGuildDBState
	SubguildNameTaken
)
