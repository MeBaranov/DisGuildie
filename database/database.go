package database

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Void interface{}

var Member Void

const (
	_ = iota
	Number
	Str
)

type Stat struct {
	ID          string
	Type        int
	Description string
}

type Guild struct {
	GuildId          uuid.UUID
	ParentId         uuid.UUID
	TopLevelParentId uuid.UUID
	DiscordId        string
	Name             string
	Stats            map[string]*Stat
	ChildNames       map[string]Void
}

type GuildPermission struct {
	TopGuild    string
	GuildId     uuid.UUID
	Permissions int
}

type User struct {
	UserId    uuid.UUID
	DiscordId string
	Guilds    map[string]*GuildPermission
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
	AddGuild(g *Guild) (*Guild, *Error)
	GetGuild(g uuid.UUID) (*Guild, *Error)
	GetGuildN(d string, n string) (*Guild, *Error)
	GetGuildD(d string) (*Guild, *Error)
	GetSubGuilds(g uuid.UUID) (map[uuid.UUID]*Guild, *Error)
	RenameGuild(g uuid.UUID, name string) (*Guild, *Error)
	MoveGuild(g uuid.UUID, parent uuid.UUID) (*Guild, *Error)
	RemoveGuild(g uuid.UUID) (*Guild, *Error)
	RemoveGuildD(d string) (*Guild, *Error)

	AddGuildStat(g uuid.UUID, s *Stat) (*Guild, *Error)
	RemoveGuildStat(g uuid.UUID, n string) (*Guild, *Error)
	RemoveAllGuildStats(g uuid.UUID) (*Guild, *Error)

	AddUser(u *User, g *GuildPermission) (*User, *Error)
	GetUserD(d string) (*User, *Error)
	GetUsersInGuild(d string) ([]*User, *Error)
	SetUserPermissions(u string, g *GuildPermission) (*User, *Error)
	SetUserSubGuild(u string, g *GuildPermission) (*User, *Error)
	RemoveUserD(d string, g string) (*User, *Error)
	EraseUserD(d string) (*User, *Error)

	AddCharacter(c *Character) (*Character, *Error)
	GetCharacters(u uuid.UUID) ([]*Character, *Error)
	GetMainCharacter(u uuid.UUID) (*Character, *Error)
	GetCharacter(u uuid.UUID, n string) (*Character, *Error)
	RenameCharacter(u uuid.UUID, old string, name string) (*Character, *Error)
	ChangeMainCharacter(u uuid.UUID, name string) (*Character, *Error)
	SetCharacterStat(u uuid.UUID, name string, s string, v interface{}) (*Character, *Error)
	ChangeCharacterOwner(old uuid.UUID, name string, u uuid.UUID) (*Character, *Error)
	RemoveCharacterStat(u uuid.UUID, name string, s string) (*Character, *Error)
	RemoveCharacter(u uuid.UUID, name string) (*Character, *Error)

	AddRole(r *Role) (*Role, *Error)
	GetRole(g string, r string) (*Role, *Error)
	GetGuildRoles(g string) ([]*Role, *Error)
	SetRolePermissions(g string, r string, p int) (*Role, *Error)
	RemoveRole(g string, r string) (*Role, *Error)

	AddMoney(m *Money) (*Money, *Error)
	GetMoney(g string) (*Money, *Error)
	ChangeMoneyOwner(g string, u string) (*Money, *Error)
	SetMoneyValid(g string, t time.Time) (*Money, *Error)

	Export() ([]byte, *Error)
	Import(b []byte) *Error
}

type ErrorCode int
type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	return e.Message
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
	IOErrorDuringImport
)

const (
	SubPerm              = 0b00
	EditSubCharsPerm     = 0b01
	EditSubStructurePerm = 0b10

	OneUpPerm              = 0x0000b
	EditOneUpCharsPerm     = 0x0100b
	EditOneUpStructurePerm = 0x1000b

	GuildPerm              = 0x000000b
	EditGuildCharsPerm     = 0x010000b
	EditGuildStructurePerm = 0x100000b

	FullPermissions      = 0x111111b
	StructurePermissions = EditSubStructurePerm | EditOneUpStructurePerm | EditGuildStructurePerm
	CharsPermissions     = EditSubCharsPerm | EditOneUpCharsPerm | EditGuildCharsPerm
)

var stringToPermission = map[string]int{
	"subedituser":  EditSubCharsPerm,
	"su":           EditSubCharsPerm,
	"subeditguild": EditSubStructurePerm,
	"sg":           EditSubStructurePerm,

	"oneupedituser":  EditOneUpCharsPerm,
	"ou":             EditOneUpCharsPerm,
	"oneupeditguild": EditOneUpStructurePerm,
	"og":             EditOneUpStructurePerm,

	"guildedituser":  EditGuildCharsPerm,
	"gu":             EditGuildCharsPerm,
	"guildeditguild": EditSubStructurePerm,
	"gg":             EditGuildStructurePerm,
}

var permToString = map[int]string{
	EditSubCharsPerm:     "SubEditUser",
	EditSubStructurePerm: "SubEditGuild",

	EditOneUpCharsPerm:     "OneUpEditUser",
	EditOneUpStructurePerm: "OneUpEditGuild",

	EditGuildCharsPerm:     "GuildEditUser",
	EditGuildStructurePerm: "GuildEditGuild",
}

var stringToType = map[string]int{
	"str": Str,
	"num": Number,
	"int": Number,
}

var typeToString = map[int]string{
	Str:    "str",
	Number: "int",
}

func StringToPermission(s string) (int, error) {
	s = strings.ToLower(s)
	if rv, ok := stringToPermission[s]; ok {
		return rv, nil
	}

	return 0, errors.New("Permission " + s + " is not defined")
}

func PermissionToString(perm int) string {
	rv, added := "", false
	for v, s := range permToString {
		if perm&v != 0 {
			if added {
				rv += ", "
			}
			rv += s
			added = true
		}
	}

	if rv == "" {
		return "None"
	}
	return rv
}

func StringToType(s string) (int, error) {
	s = strings.ToLower(s)
	if rv, ok := stringToType[s]; ok {
		return rv, nil
	}

	return 0, errors.New("Type " + s + " is not defined")
}

func TypeToString(t int) string {
	if rv, ok := typeToString[t]; ok {
		return rv
	}

	return "undefined"
}
