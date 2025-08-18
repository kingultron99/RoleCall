package events

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"kingultron99.com/RoleCall/core"
	"kingultron99.com/RoleCall/interactions"
	"kingultron99.com/RoleCall/utils"
)

func AddJoinHandler() {
	core.State.AddHandler(func(e *gateway.GuildCreateEvent) {
		_, err := core.DB.Exec("INSERT INTO guilds(guild_id) VALUES($1)", e.ID)
		if err != nil {
			log.Print("Failed to insert guild into DB!")
		}
	})

	// Listen for user joins.
	// if they are joining a server with member screening isPending will = true
	// add them to the pending table. Otherwise, assign the role.
	core.State.AddHandler(func(e *gateway.GuildMemberAddEvent) {
		row := core.DB.QueryRow("SELECT role_id FROM autoroles WHERE guild_id=$1", e.GuildID.String())
		var role string
		noRole := row.Scan(&role)
		if noRole == sql.ErrNoRows {
			return
		}

		if e.IsPending {
			core.DB.Exec("INSERT INTO pending_users(guild_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", e.GuildID.String(), e.User.ID.String())
			return
		}

		err := core.State.AddRole(e.GuildID, e.User.ID, discord.RoleID(utils.MustSnowflakeEnv(role)), api.AddRoleData{})
		if err != nil {
			log.Print(err)
		}
	})

	// Wait for user updates and check if the user finished
	// the screening (isPending becomes false), and assign role
	core.State.AddHandler(func(e *gateway.GuildMemberUpdateEvent) {
		userRow := core.DB.QueryRow("SELECT user_id FROM pending_users WHERE guild_id=$1", e.GuildID.String())
		var user string
		err := userRow.Scan(&user)
		if err == sql.ErrNoRows {
			return
		}

		roleRow := core.DB.QueryRow("SELECT role_id FROM autoroles WHERE guild_id=$1", e.GuildID.String())
		var role string
		noRole := roleRow.Scan(&role)
		// This technically shouldnt happen, as the user shouldnt be added to the pending list if the server doesnt have their default role configured
		if noRole == sql.ErrNoRows {
			return
		}

		if user == e.User.ID.String() && !e.IsPending {
			err := core.State.AddRole(e.GuildID, e.User.ID, discord.RoleID(utils.MustSnowflakeEnv(role)), api.AddRoleData{})
			if err != nil {
				log.Print(err)
			}
			core.DB.Exec("DELETE FROM pending_users WHERE guild_id=$1 AND user_id=$2", e.GuildID.String(), e.User.ID.String())
		}
	})

	// Remove the user from the pending list if they leave
	// before completing the server screening
	core.State.AddHandler(func(e *gateway.GuildMemberRemoveEvent) {
		userRow := core.DB.QueryRow("SELECT user_id FROM pending_users WHERE guild_id=$1", e.GuildID.String())
		var user string
		err := userRow.Scan(&user)
		if err == sql.ErrNoRows {
			return
		}
		core.DB.Exec("DELETE FROM pending_users WHERE guild_id=$1 AND user_id=$2", e.GuildID.String(), e.User.ID.String())
	})

	core.State.AddHandler(func(e *gateway.GuildDeleteEvent) {
		// We havent been removed from the guild, the guild has just become unavailable due to an outage...
		if e.Unavailable {
			return
		}
		core.DB.Exec("DELETE FROM guilds WHERE guild_id=$1", e.ID.String())
	})
}

func RegisterCommands(appID discord.AppID, guildID discord.GuildID) {

	var commands []api.CreateCommandData

	for _, command := range interactions.MapCommands {
		if command.Exclude != true && command.CreateCommandData != nil {
			commands = append(commands, *command.CreateCommandData)
		}
	}

	// Register commands as guild commands for testing
	if *core.DevPtr && guildID.String() == os.Getenv("Testing_Guild") {
		_, err := core.State.BulkOverwriteGuildCommands(appID, guildID, commands)
		if err != nil {
			log.Printf("Failed to overwrite commands in %v with err: %v", guildID, err)
		}
		log.Print("Commands registered to guilds")
		return
	}

	_, err := core.State.BulkOverwriteCommands(appID, commands)
	if err != nil {
		log.Printf("Failed to overwrite commands with err: %v", err)
	}
	log.Print("Commands registered globally")

}

func CommandRouter() {
	core.State.AddHandler(func(e *gateway.InteractionCreateEvent) {
		switch data := e.Data.(type) {
		case *discord.CommandInteraction:
			go interactions.MapCommands[data.Name].Run(e, data)
		case *discord.StringSelectInteraction:
			go interactions.MapStringComponents[strings.Split(string(data.CustomID), "_")[0]].Run(e, data)
		case *discord.RoleSelectInteraction:
			go interactions.MapRoleComponents[strings.Split(string(data.CustomID), "_")[0]].Run(e, data)
		case *discord.ButtonInteraction:
			go interactions.MapButtonComponents[strings.Split(string(data.CustomID), "_")[0]].Run(e, data)
		case *discord.ModalInteraction:
			go interactions.MapModalComponents[strings.Split(string(data.CustomID), "_")[0]].Run(e, data)
		case *discord.ChannelSelectInteraction:
			go interactions.MapChannelComponents[strings.Split(string(data.CustomID), "_")[0]].Run(e, data)
		default:
			go core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.MessageInteractionWithSource,
				Data: &api.InteractionResponseData{
					Flags: discord.EphemeralMessage,
					Embeds: &[]discord.Embed{
						{Title: "Uh...", Description: "We dont know how to handle that interaction yet!", Color: utils.Red},
					},
				},
			})
		}
	})
	log.Print("Command router successfully registered")
}
