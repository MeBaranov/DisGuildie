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
	DefaultStat      string
	StatVersion      int
}

type GuildPermission struct {
	TopGuild    string
	GuildId     uuid.UUID
	Permissions int
}

type User struct {
	Id     string
	Guilds map[string]*GuildPermission
}

type Character struct {
	GuildId     string
	UserId      string
	Name        string
	Main        bool
	Body        map[string]interface{}
	StatVersion int
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
	GetGuildN(d string, n string) (*Guild, error)
	GetGuildD(d string) (*Guild, error)
	GetSubGuilds(g uuid.UUID) (map[uuid.UUID]*Guild, error)
	RenameGuild(g uuid.UUID, name string) (*Guild, error)
	MoveGuild(g uuid.UUID, parent uuid.UUID) (*Guild, error)
	RemoveGuild(g uuid.UUID) (*Guild, error)
	RemoveGuildD(d string) (*Guild, error)

	AddGuildStat(g uuid.UUID, s *Stat) (*Guild, error)
	SetDefaultGuildStat(g uuid.UUID, sn string) (*Guild, error)
	RemoveGuildStat(g uuid.UUID, n string) (*Guild, error)
	RemoveAllGuildStats(g uuid.UUID) (*Guild, error)

	AddUser(d string, g *GuildPermission) (*User, error)
	GetUserD(d string) (*User, error)
	GetUsersInGuild(d string) ([]*User, error)
	SetUserPermissions(u string, g *GuildPermission) (*User, error)
	SetUserSubGuild(u string, g *GuildPermission) (*User, error)
	RemoveUserD(d string, g string) (*User, error)
	EraseUserD(d string) (*User, error)

	AddCharacter(c *Character) (*Character, error)
	GetCharacters(g string, u string) ([]*Character, error)
	GetCharactersSorted(g string, s string, t int, asc bool, limit int) ([]*Character, error)
	GetCharactersOutdated(g string, v int) ([]*Character, error)
	GetCharactersByName(g string, n string) ([]*Character, error)
	GetMainCharacter(g string, u string) (*Character, error)
	GetCharacter(g string, u string, n string) (*Character, error)
	RenameCharacter(g string, u string, old string, name string) (*Character, error)
	ChangeMainCharacter(g string, u string, name string) (*Character, error)
	SetCharacterStat(g string, u string, name string, s string, v interface{}) (*Character, error)
	SetCharacterStatVersion(g string, u string, name string, stats map[string]*Stat, version int) (*Character, error)
	ChangeCharacterOwner(g string, old string, name string, u string) (*Character, error)
	RemoveCharacterStat(g string, u string, name string, s string) (*Character, error)
	RemoveCharacter(g string, u string, name string) (*Character, error)

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
	UnknownStatType
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

func ErrToDbErr(e error) *Error {
	if rv, ok := e.(*Error); ok {
		return rv
	}

	return nil
}
