package interactions

import (
	"database/sql"
	"strconv"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"kingultron99.com/RoleCall/core"
	"kingultron99.com/RoleCall/utils"
)

func init() {
	MapCommands["delete"] = Command{
		CreateCommandData: &api.CreateCommandData{
			Name:        "delete",
			Description: "Commands to remove your data from RoleCall",
			Options: discord.CommandOptions{
				&discord.SubcommandOption{
					OptionName:  "everything",
					Description: "Deletes all configurations related to your guild from RoleCall",
				},
				&discord.SubcommandOption{
					OptionName:  "providers",
					Description: "Deletes one or more of your configured provider messages and their associated records",
				},
				&discord.SubcommandOption{
					OptionName:  "autorole",
					Description: "Deletes your autorole configuration",
				},
				&discord.SubcommandOption{
					OptionName:  "leftovers",
					Description: "Deletes any entries that may have been left over from a misconfigured role provider",
				},
				&discord.SubcommandOption{
					OptionName:  "comms",
					Description: "Resets your broadcast preferences to default",
				},
			},
			DefaultMemberPermissions: discord.NewPermissions(discord.PermissionAdministrator),
		},
		Run: func(e *gateway.InteractionCreateEvent, data *discord.CommandInteraction) {
			switch data.Options[0].Name {
			case "everything":
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title: "Are you absolutely sure?",
								Description: `Running this command will **permanently** delete all configurations from RoleCall.
Your guild ID will remain saved until RoleCall is removed from the guild
Any messages sent by RoleCall will also be deleted.
								
This will not affect any of your roles, and any users who have given themselves roles via RoleCall will continue to have those roles.

This will *not* remove your guilds entry from RoleCalls guild list, UNTIL RoleCall is removed from your server.

This will also set also set your broadcast preferences back to default (i.e. all messages enabled, and no custom channel selected)
								
### This is your only warning. Choose carefully.`,
								Color: utils.Red,
							},
						},
						Components: &discord.ContainerComponents{
							&discord.ActionRowComponent{
								&discord.ButtonComponent{
									Style:    discord.DangerButtonStyle(),
									Label:    "Proceed.",
									CustomID: "deleteall",
								},
								&discord.ButtonComponent{
									Style:    discord.SecondaryButtonStyle(),
									Label:    "Nevermind...",
									CustomID: "canceldelete",
								},
							},
						},
					},
				})
			case "providers":
				messages, _ := core.DB.Query("SELECT id, channel_id FROM interaction_messages WHERE guild_id=$1", e.GuildID)

				options := []discord.SelectOption{}

				// sql.ErrNoRows is only given when using `QueryRow`.
				// Query will just return an empty `*sql.rows` and `.Next()` will be false by default
				// using this, we can make our own little checker like so:
				found := false
				for messages.Next() {
					// will only be evaluated when `.Next()` is true, and the for loop executes
					found = true
					var message InteractionMsg
					messages.Scan(&message.ID, &message.ChannelID)

					res := core.DB.QueryRow("SELECT COUNT(*) FROM interaction_roles WHERE interaction_msg_id=$1", message.ID)
					var roleCount string
					res.Scan(&roleCount)

					channel, _ := core.State.Channel(discord.ChannelID(utils.MustSnowflakeEnv(message.ChannelID)))

					options = append(options, discord.SelectOption{
						Label:       "Provider in " + channel.Name,
						Description: roleCount + " associated roles",
						Value:       strconv.Itoa(message.ID),
					})
				}

				if !found {
					core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
						Type: api.MessageInteractionWithSource,
						Data: &api.InteractionResponseData{
							Flags: discord.EphemeralMessage,
							Embeds: &[]discord.Embed{
								{
									Title:       "Must've been the wind...",
									Description: "Your guild doesnt have any active role providers configured!",
									Color:       utils.Red,
								},
							},
						},
					})
					return
				}

				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "Current providers",
								Description: "Pick what provider messages you want to remove",
								Color:       utils.Purple,
							},
						},
						Components: &discord.ContainerComponents{
							&discord.ActionRowComponent{
								&discord.StringSelectComponent{
									Options:     options,
									CustomID:    "deleteprovider",
									ValueLimits: [2]int{1, len(options)},
									Placeholder: "Select messages...",
								},
							},
						},
					},
				})
			case "autorole":
				core.DB.Exec("DELETE FROM autoroles WHERE guild_id=$1", e.GuildID)
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "Purge complete",
								Description: "We've deleted your autorole configuration!",
								Color:       utils.Green,
							},
						},
						Components: &discord.ContainerComponents{},
					},
				})
			case "leftovers":
				core.DB.Exec("DELETE FROM interaction_roles WHERE interaction_msg_id IS NULL")
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "Purge complete",
								Description: "We've deleted any roles that werent associated with a message.",
								Color:       utils.Green,
							},
						},
						Components: &discord.ContainerComponents{},
					},
				})
			case "comms":
				core.DB.Exec("UPDATE broadcast_preferences SET system=true, updates=true, location=NULL WHERE guild_id=$1;", e.GuildID)
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "Broadcast preferences reset",
								Description: "We've reset your broadcast preferences to default (i.e. all messages enabled, and no custom channel selected)!",
								Color:       utils.Green,
							},
						},
						Components: &discord.ContainerComponents{},
					},
				})
			}
		},
	}

	MapButtonComponents["canceldelete"] = ButtonComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ButtonInteraction) {
			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Embeds: &[]discord.Embed{
						{
							Title:       "Phew... That was a close one",
							Description: "Deletion cancelled!\n\n-# You can dismiss this message safely",
							Color:       utils.Green,
						},
					},
					Components: &discord.ContainerComponents{},
				},
			})
		},
	}
	MapButtonComponents["deleteall"] = ButtonComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ButtonInteraction) {

			rows, _ := core.DB.Query("SELECT message_id, channel_id FROM interaction_messages WHERE guild_id=$1", e.GuildID)

			for rows.Next() {
				var provider InteractionMsg
				rows.Scan(&provider.MessageID, &provider.ChannelID)
				core.State.DeleteMessage(discord.ChannelID(utils.MustSnowflakeEnv(provider.ChannelID)), discord.MessageID(utils.MustSnowflakeEnv(provider.MessageID)), api.AuditLogReason("Purging RoleCall configurations & messages"))
			}

			// Delete related records
			// Deleting from interaction_messages will cascade to interaction_roles (i.e. deleting an interaction_message will delete all associated interaction_roles)
			core.DB.Exec("DELETE FROM interaction_messages WHERE guild_id=$1;", e.GuildID)
			// However, I fully expect people not to read warnings, and will leave orphaned entries in interaction_roles without a message
			core.DB.Exec("DELETE FROM interaction_roles WHERE guild_id=$1;", e.GuildID)
			core.DB.Exec("DELETE FROM autoroles WHERE guild_id=$1;", e.GuildID)
			core.DB.Exec("DELETE FROM pending_users WHERE guild_id=$1;", e.GuildID)
			core.DB.Exec("UPDATE broadcast_preferences SET system=true, updates=true, location=NULL WHERE guild_id=$1;", e.GuildID)

			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Embeds: &[]discord.Embed{
						{
							Title:       "Purge complete",
							Description: "Sorry to see you go...\n\nWe've deleted all recorded configurations and associated messages.",
							Color:       utils.Green,
						},
					},
					Components: &discord.ContainerComponents{},
				},
			})
		},
	}

	MapStringComponents["deleteprovider"] = StringSelectComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.StringSelectInteraction) {
			for _, id := range data.Values {
				row := core.DB.QueryRow("SELECT channel_id, message_id FROM interaction_messages WHERE id=$1", id)
				var msg InteractionMsg
				err := row.Scan(&msg.ChannelID, &msg.MessageID)
				if err == sql.ErrNoRows {
					core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
						Type: api.UpdateMessage,
						Data: &api.InteractionResponseData{
							Embeds: &[]discord.Embed{
								{
									Title:       "That message doesn't exist!",
									Description: "It looks like you've somehow selected an option from a leftover command!\n\nPlease dismiss this message and try again",
									Color:       utils.Red,
								},
							},
							Components: &discord.ContainerComponents{},
						},
					})
				}

				core.State.DeleteMessage(discord.ChannelID(utils.MustSnowflakeEnv(msg.ChannelID)), discord.MessageID(utils.MustSnowflakeEnv(msg.MessageID)), api.AuditLogReason(e.Member.User.Username+" deleted provider"))
				core.DB.Exec("DELETE FROM interaction_messages WHERE id=$1", id)
			}
			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Embeds: &[]discord.Embed{
						{
							Title:       "Purge complete",
							Description: "We've deleted all recorded configurations and associated messages.\n\nYour broadcast preferences have been reset to default.",
							Color:       utils.Green,
						},
					},
					Components: &discord.ContainerComponents{},
				},
			})
		},
	}

}
