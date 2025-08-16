package interactions

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

var MapCommands = make(map[string]Command)
var MapStringComponents = make(map[string]StringSelectComponent)
var MapRoleComponents = make(map[string]RoleSelectComponent)
var MapChannelComponents = make(map[string]ChannelSelectComponent)
var MapButtonComponents = make(map[string]ButtonComponent)
var MapModalComponents = make(map[string]ModalComponent)

type Command struct {
	*api.CreateCommandData
	Exclude bool // should this command be excluded from registration?
	Run     func(e *gateway.InteractionCreateEvent, data *discord.CommandInteraction)
}
type StringSelectComponent struct {
	Run func(e *gateway.InteractionCreateEvent, data *discord.StringSelectInteraction)
}
type RoleSelectComponent struct {
	Run func(e *gateway.InteractionCreateEvent, data *discord.RoleSelectInteraction)
}
type ChannelSelectComponent struct {
	Run func(e *gateway.InteractionCreateEvent, data *discord.ChannelSelectInteraction)
}
type ButtonComponent struct {
	Run func(e *gateway.InteractionCreateEvent, data *discord.ButtonInteraction)
}
type ModalComponent struct {
	Run func(e *gateway.InteractionCreateEvent, data *discord.ModalInteraction)
}
type InteractionRole struct {
	ID               int
	GuildID          string // the guild ID
	Label            string // the text on the button
	RoleID           string // ID of the role we're applying
	Password         string // password protection for the role
	Uses             int    // how many successful times has this been used?
	InteractionMsgID int
}

type InteractionMsg struct {
	ID        int
	GuildID   string
	ChannelID string
	MessageID string
}
