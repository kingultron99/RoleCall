package interactions

import (
	"database/sql"
	"log"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"kingultron99.com/RoleCall/core"
	"kingultron99.com/RoleCall/utils"
)

func init() {
	MapCommands["configure"] = Command{
		CreateCommandData: &api.CreateCommandData{
			Name:                     "configure",
			Description:              "Create a new role giver",
			NoDMPermission:           true,
			Type:                     discord.ChatInputCommand,
			DefaultMemberPermissions: discord.NewPermissions(discord.PermissionAdministrator),
			Options: discord.CommandOptions{
				&discord.SubcommandOption{
					OptionName:  "provider",
					Description: "Creates a new role provider",
				},
				&discord.SubcommandOption{
					OptionName:  "autorole",
					Description: "Configure a default role that all new members will be given",
				},
				&discord.SubcommandGroupOption{
					OptionName:  "broadcasts",
					Description: "Confifgure broadcast messages from RoleCall",
					Subcommands: []*discord.SubcommandOption{
						{
							OptionName:  "system",
							Description: "Toggle system status messages for when RoleCall is having issues or is undergoing maintenance",
							Options: []discord.CommandOptionValue{
								&discord.BooleanOption{
									OptionName:  "enabled",
									Description: "Whether or not to receive system notice messages from RoleCall",
									Required:    true,
								},
							},
						},
						{
							OptionName:  "updates",
							Description: "Toggle update messages that report when RoleCall has new features or improvements",
							Options: []discord.CommandOptionValue{
								&discord.BooleanOption{
									OptionName:  "enabled",
									Description: "Whether or not to receive update messages from RoleCall",
									Required:    true,
								},
							},
						},
						{
							OptionName:  "location",
							Description: "Configure where you receive these messages",
							Options: []discord.CommandOptionValue{
								&discord.ChannelOption{
									OptionName:  "channel",
									Description: "The target channel for RoleCall's broadcast messages",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
		Run: func(e *gateway.InteractionCreateEvent, data *discord.CommandInteraction) {
			switch data.Options[0].Name {
			case "provider":
				me, _ := core.State.Me()
				mem, _ := core.State.Member(e.GuildID, me.ID)

				embed := discord.Embed{
					Title: "Pick your Roles!",
					Description: `Select **Up to 5** roles that you would like this message to offer
						
Just remember, RoleCall cannot give anyone the <@&` + mem.RoleIDs[0].String() + `> role or higher!`,
					Color: utils.Purple,
				}
				row := core.DB.QueryRow("SELECT * FROM interaction_messages WHERE guild_id=$1", e.GuildID)
				err := row.Scan()

				if err == sql.ErrNoRows {
					embed = discord.Embed{
						Title: "Pick your Roles!",
						Description: `# Important!
Hey! It looks like this is your first time setting up one of our role providers!
					
Please make sure that you drag <@&` + mem.RoleIDs[0].String() + `> **ABOVE** all other roles that you want RoleCall to be able to manage in your guild settings!
					
RoleCall cannot give anyone the <@&` + mem.RoleIDs[0].String() + `> role or higher!

Select **Up to 5** roles that you would like this message to offer`,
						Color: utils.Purple,
					}
				}
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Embeds: &[]discord.Embed{embed},
						Flags:  discord.EphemeralMessage,
						Components: &discord.ContainerComponents{
							&discord.ActionRowComponent{
								&discord.RoleSelectComponent{
									CustomID:    discord.ComponentID("roles_" + data.GuildID.String()),
									Placeholder: "Search your roles...",
									ValueLimits: [2]int{1, 5},
								},
							},
						},
					},
				})
			case "autorole":
				me, _ := core.State.Me()
				mem, _ := core.State.Member(e.GuildID, me.ID)

				embed := discord.Embed{
					Title: "Select your default role",
					Description: `This role will be given to anyone that joins your guild.
					
You can change this role at any time just by running this command again!
					
RoleCall cannot give anyone the <@&` + mem.RoleIDs[0].String() + `> role or higher!`,
					Color: utils.Purple,
				}

				row := core.DB.QueryRow("SELECT * FROM autoroles WHERE guild_id=$1", e.GuildID)
				err := row.Scan()

				if err == sql.ErrNoRows {
					embed = discord.Embed{
						Title: "Select your default role",
						Description: `# Important!
Hey! It looks like this is your first time setting up RoleCall's autorole feature.
						
Please make sure that you drag <@&` + mem.RoleIDs[0].String() + `> **ABOVE** all other roles that you want RoleCall to be able to manage in your server settings!
						
This role will be given to anyone that joins your guild.
						
You can change this role at any time just by running this command again!`,
						Color: utils.Purple,
					}
				}
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags:  discord.EphemeralMessage,
						Embeds: &[]discord.Embed{embed},
						Components: &discord.ContainerComponents{
							&discord.ActionRowComponent{
								&discord.RoleSelectComponent{
									CustomID:    "autorole",
									Placeholder: "search roles...",
								},
							},
						},
					},
				})
			case "broadcasts":
				switch data.Options[0].Options[0].Name {
				case "system":
					enabled, _ := data.Options[0].Options[0].Options[0].BoolValue()
					_, err := core.DB.Exec("UPDATE broadcast_preferences SET system=$1 WHERE guild_id=$2", enabled, e.GuildID)
					if err != nil {
						log.Println(err)
						core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
							Type: api.MessageInteractionWithSource,
							Data: &api.InteractionResponseData{
								Flags: discord.EphemeralMessage,
								Embeds: &[]discord.Embed{
									{
										Color:       utils.Red,
										Title:       "Something went wrong!!",
										Description: "Your preferences failed to update. Please try again later, and if this problem persists, please join the support server and let us know!",
									},
								},
							},
						})
					}

					var messageType string
					if enabled {
						messageType = "Starting now, you'll receive system notices from RoleCall!"
					} else {
						messageType = "You'll no longer receive system notices from RoleCall. If you change your mind, you can re-enable them with the `/configure broadcasts` command!"
					}

					core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
						Type: api.MessageInteractionWithSource,
						Data: &api.InteractionResponseData{
							Flags: discord.EphemeralMessage,
							Embeds: &[]discord.Embed{
								{
									Color:       utils.Green,
									Title:       "All set!",
									Description: "Your preferences been updated successfully!\n\n" + messageType,
								},
							},
						},
					})
				case "updates":
					enabled, _ := data.Options[0].Options[0].Options[0].BoolValue()
					core.DB.Exec("UPDATE broadcast_preferences SET updates=$1 WHERE guild_id=$2", enabled, e.GuildID)

					var messageType string
					if enabled {
						messageType = "Starting now, you'll receive update messages from RoleCall!"
					} else {
						messageType = "You will no longer receive update messages from RoleCall. If you change your mind, you can re-enable them with the `/configure broadcasts` command!"
					}

					core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
						Type: api.MessageInteractionWithSource,
						Data: &api.InteractionResponseData{
							Flags: discord.EphemeralMessage,
							Embeds: &[]discord.Embed{
								{
									Title:       "All set!",
									Description: "Your preferences been updated successfully!\n\n" + messageType,
								},
							},
						},
					})
				case "location":
					channelID := data.Options[0].Options[0].Options[0].Value.String()
					channelID = strings.Trim(channelID, "\"")

					_, err := core.DB.Exec("UPDATE broadcast_preferences SET location=$1 WHERE guild_id=$2", utils.MustSnowflakeEnv(channelID), e.GuildID)
					if err != nil {
						log.Println(err)
					}

					core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
						Type: api.MessageInteractionWithSource,
						Data: &api.InteractionResponseData{
							Flags: discord.EphemeralMessage,
							Embeds: &[]discord.Embed{
								{
									Title:       "Your preferences been updated successfully!",
									Description: "From now on, all broadcast messages from RoleCall will be sent in <#" + channelID + ">!\n\n**Please Make sure that RoleCall has permission to __view__ and __send messages__ the target channel!**",
									Color:       utils.Green,
								},
							},
						},
					})
				}
			}
		},
	}
	MapRoleComponents["roles"] = RoleSelectComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.RoleSelectInteraction) {

			var comp []discord.Component

			var buttonRow = make(discord.ActionRowComponent, len(data.Values))

			comp = append(comp, &discord.ActionRowComponent{
				&discord.ButtonComponent{
					Style:    discord.PrimaryButtonStyle(),
					CustomID: "editmessage",
					Label:    "Edit message",
				},
				&discord.ButtonComponent{
					Label:    "Done!",
					Style:    discord.SuccessButtonStyle(),
					CustomID: "editnewfinish",
				},
				&discord.ButtonComponent{
					Label:    "Cancel!",
					Style:    discord.DangerButtonStyle(),
					CustomID: "editnewcancel",
				},
			})

			// Get all the roles from the current guild
			guildRoles, _ := core.State.Roles(e.GuildID)
			isRoleManagedMap := make(map[discord.RoleID]bool)

			// iterate through all the guilds roles and map which roles are managed
			for _, role := range guildRoles {
				isRoleManagedMap[role.ID] = role.Managed
			}

			roles := []discord.RoleID{}

			for i, role := range data.Values {
				if isRoleManagedMap[role] {
					core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
						Type: api.MessageInteractionWithSource,
						Data: &api.InteractionResponseData{
							Flags: discord.EphemeralMessage,
							Embeds: &[]discord.Embed{{
								Title:       "Woah!",
								Description: "<@&" + role.String() + "> is managed by RoleCall or another integration!\n\nWe can't assign managed roles to users!",
								Color:       utils.Red,
							}},
						},
					})
					return
				}

				target, _ := core.State.Role(e.GuildID, role)
				// Roles can be up to 100 characters, but buttons only allow up to 80!
				label := target.Name
				if len(label) > 80 {
					label = strings.TrimRight(label[:77], " ") + "..."
				}
				roles = append(roles, target.ID)
				buttonRow[i] = &discord.ButtonComponent{
					Style:    discord.SecondaryButtonStyle(),
					CustomID: discord.ComponentID("rename_" + role.String()),
					Label:    label,
				}

				row := core.DB.QueryRow("SELECT role_id FROM interaction_roles WHERE guild_id=$1 AND role_id=$2", e.GuildID, role)

				var existing InteractionRole

				if err := row.Scan(&existing.RoleID); err != sql.ErrNoRows {
					core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
						Type: api.MessageInteractionWithSource,
						Data: &api.InteractionResponseData{
							Flags: discord.EphemeralMessage,
							Embeds: &[]discord.Embed{{
								Title:       "Error!",
								Description: "One or more of your selected roles have *already been configured*!\n\nOffending role: <@&" + existing.RoleID + ">\n\nIf you're seeing this message and the target role is infact NOT on any of your existing role providers (if any), please run `/delete leftovers` to clean up the erroneous entries, and try again.",
								Color:       utils.Red,
							}},
						},
					})
					return
				}
			}

			// Finally dump the result into the DB
			for _, role := range roles {
				core.DB.Exec("INSERT INTO interaction_roles(guild_id,role_id) VALUES ($1,$2)", e.GuildID, role)
			}

			comp = append(comp, &buttonRow)

			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Content: option.NewNullableString("Select your roles:"),
					Embeds: &[]discord.Embed{{
						Title: "THIS IS A PREVIEW",
						Description: `Click any of the grey buttons to edit their label and password.
						
This will be how they appear on the actual message!
						
Please do **not** dismiss this message, this will leave residual entries in the database! 
If you want to cancel, press the cancel button, this will automatically remove those entries.`,
						Color: utils.Purple,
					}},
					Components: discord.ComponentsPtr(comp...),
				},
			})
		},
	}
	// Renaming the button lables
	MapButtonComponents["rename"] = ButtonComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ButtonInteraction) {

			target, _ := core.State.Role(e.GuildID, discord.RoleID(utils.MustSnowflakeEnv(strings.Split(string(data.CustomID), "_")[1])))

			title := target.Name

			// Prevent discord from screaming at me because the modal title is too long
			if len(title) > 26 {
				title = strings.TrimRight(title[:23], " ") + "..."
			}

			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.ModalResponse,
				Data: &api.InteractionResponseData{
					CustomID: option.NewNullableString(string(data.CustomID)),
					Title:    option.NewNullableString("Editing button for " + title),
					Components: &discord.ContainerComponents{
						&discord.ActionRowComponent{
							&discord.TextInputComponent{
								CustomID:     "content",
								Label:        "Button Label",
								LengthLimits: [2]int{0, 80},
								Style:        discord.TextInputShortStyle,
								Placeholder:  "What should the button say?",
								Value:        e.Message.Components.Find(discord.ComponentID(data.CustomID)).(*discord.ButtonComponent).Label,
								Required:     true,
							},
						},
						&discord.ActionRowComponent{
							&discord.TextInputComponent{
								CustomID:     "password",
								Label:        "Button Password",
								LengthLimits: [2]int{0, 100},
								Style:        discord.TextInputShortStyle,
								Placeholder:  "Leave blank for no password.",
								Required:     false,
							},
						},
					},
				},
			})
		},
	}
	MapModalComponents["rename"] = ModalComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ModalInteraction) {
			label := data.Components.Find("content").(*discord.TextInputComponent).Value
			var password = ""

			if data.Components.Find("password").(*discord.TextInputComponent).Value != "" {
				password, _ = utils.HashPassword(data.Components.Find("password").(*discord.TextInputComponent).Value)
			}

			id := strings.Split(string(data.CustomID), "_")[1]

			var newComponents = make(discord.ContainerComponents, 2)

			for i, actionRow := range e.Message.Components {
				// we know the first layer are all actionrows. so we can do some typecasting to figure out their length
				row, _ := actionRow.(*discord.ActionRowComponent)
				newRow := make(discord.ActionRowComponent, len(*row))

				for i, component := range *row {
					if string(component.ID()) != string(data.CustomID) {
						newRow[i] = component
						continue
					}
					newRow[i] = &discord.ButtonComponent{
						Style:    discord.SecondaryButtonStyle(),
						CustomID: data.CustomID,
						Label:    label,
					}
				}
				newComponents[i] = &newRow
			}

			core.DB.Exec("UPDATE interaction_roles SET  password=$1 WHERE role_id=$2 AND guild_id = $3", password, id, e.GuildID)

			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Content:    option.NewNullableString(e.Message.Content),
					Embeds:     &e.Message.Embeds,
					Components: &newComponents,
				},
			})
		},
	}

	// Editing the message content
	MapButtonComponents["editmessage"] = ButtonComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ButtonInteraction) {
			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.ModalResponse,
				Data: &api.InteractionResponseData{
					CustomID: option.NewNullableString("editmessage"),
					Title:    option.NewNullableString("Editing the message"),
					Components: &discord.ContainerComponents{
						&discord.ActionRowComponent{
							&discord.TextInputComponent{
								CustomID:     "content",
								Label:        "Message",
								LengthLimits: [2]int{0, 2000},
								Style:        discord.TextInputParagraphStyle,
								Placeholder:  "An optional message to include for your role selector",
								Required:     false,
								Value:        e.Message.Content,
							},
						},
					},
				},
			})
		},
	}
	MapModalComponents["editmessage"] = ModalComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ModalInteraction) {
			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Flags:      discord.EphemeralMessage,
					Content:    option.NewNullableString(data.Components.Find("content").(*discord.TextInputComponent).Value),
					Embeds:     &e.Message.Embeds,
					Components: &e.Message.Components,
				},
			})
		},
	}

	// Cancel action
	MapButtonComponents["editnewcancel"] = ButtonComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ButtonInteraction) {

			rows, _ := core.DB.Query("SELECT role_id FROM interaction_roles WHERE guild_id=$1", e.GuildID)
			for rows.Next() {
				var row InteractionRole
				rows.Scan(&row.RoleID)

				comp := e.Message.Components.Find(discord.ComponentID("rename_" + row.RoleID))

				if comp != nil {
					core.DB.Exec("DELETE FROM interaction_roles WHERE role_id=$1 AND guild_id=$2", row.RoleID, e.GuildID)
				}

			}
			rows.Close()

			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Content: option.NewNullableString(""),
					Embeds: &[]discord.Embed{
						{
							Title:       "Setup cancelled!",
							Description: "We've removed all preliminary DB entries related to this setup.\n\nIf you want to restart the setup, simply run `/configure provider`\n-# You can safely dismiss this message now!",
							Color:       utils.Green,
						},
					},
					Components: &discord.ContainerComponents{},
				},
			})
		},
	}

	// Done action
	MapButtonComponents["editnewfinish"] = ButtonComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ButtonInteraction) {

			var newComponents = make(discord.ContainerComponents, 2)

			newComponents[0] = &discord.ActionRowComponent{&discord.ChannelSelectComponent{
				CustomID:    "target",
				Placeholder: "Target channel",
			}}

			row, _ := e.Message.Components[1].(*discord.ActionRowComponent)
			newRow := make(discord.ActionRowComponent, len(*row)+1)
			for i, component := range *row {
				newRow[i] = &discord.ButtonComponent{
					Style:    discord.SecondaryButtonStyle(),
					CustomID: component.ID(),
					Label:    component.(*discord.ButtonComponent).Label,
					Disabled: true,
				}
			}

			newRow[len(*row)] = &discord.ButtonComponent{
				Style:    discord.DangerButtonStyle(),
				CustomID: "editnewcancel",
				Label:    "Cancel!",
				Disabled: false,
			}

			newComponents[1] = &newRow

			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Flags:   discord.EphemeralMessage,
					Content: option.NewNullableString(e.Message.Content),
					Embeds: &[]discord.Embed{
						{
							Title:       "Last step!",
							Description: "Select the channel you want the final message to be sent\n\n## Make sure RoleCall has permission to send messages in the target channel BEFORE selecting it!\n\nPlease do **not** dismiss this message, this will leave residual entries in the database! If you want to cancel, press the cancel button, this will automatically remove those entries.",
							Color:       utils.Purple,
						},
					},
					Components: &newComponents,
				},
			})
		},
	}
	MapChannelComponents["target"] = ChannelSelectComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ChannelSelectInteraction) {

			c, _ := core.State.Channel(data.Values[0])

			if c.Type != discord.GuildText {
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "I can't do that...",
								Description: "I can't send messages in that kind of channel!",
								Color:       utils.Red,
							},
						},
					},
				})
				return
			}
			if c.SelfPermissions.Has(discord.PermissionSendMessages) {
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "I can't do that...",
								Description: "I dont have permission to send messages in that channel!",
								Color:       utils.Red,
							},
						},
					},
				})
				return
			}

			components := make(discord.ContainerComponents, 1)

			row, _ := e.Message.Components[1].(*discord.ActionRowComponent)
			newRow := make(discord.ActionRowComponent, len(*row)-1)

			//used later for updating each button in the DB with the final
			buttons := make([]discord.ButtonComponent, len(*row)-1)

			for i, comp := range *row {
				button := *comp.(*discord.ButtonComponent)
				if button.CustomID == "editnewcancel" {
					continue
				}
				newRow[i] = &discord.ButtonComponent{
					Label:    button.Label,
					CustomID: discord.ComponentID("gib_" + strings.Split(string(button.CustomID), "_")[1]),
					Style:    discord.SecondaryButtonStyle(),
					Disabled: false,
				}
				buttons[i] = discord.ButtonComponent{
					Label:    button.Label,
					CustomID: discord.ComponentID("gib_" + strings.Split(string(button.CustomID), "_")[1]),
					Style:    discord.SecondaryButtonStyle(),
					Disabled: false,
				}
			}
			components[0] = &newRow

			msg, err := core.State.SendMessageComplex(discord.ChannelID(data.Values[0]), api.SendMessageData{
				Content:    e.Message.Content,
				Components: components,
			})

			if err != nil {
				log.Print(err)
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "thats strange...",
								Description: "RoleCall wasnt able to send the provider to your selected channel, this is probably a bug.\nPlease contact support and provide this error:\n```" + err.Error() + "```",
								Color:       utils.Red,
							},
						},
					},
				})
				return
			}

			ret := core.DB.QueryRow("INSERT INTO interaction_messages(guild_id, channel_id, message_id) VALUES ($1,$2,$3) RETURNING id", e.GuildID, msg.ChannelID, msg.ID)

			var res InteractionMsg
			ret.Scan(&res.ID)

			for _, button := range buttons {
				core.DB.Exec("UPDATE interaction_roles SET interaction_msg_id=$1 WHERE guild_id=$2 AND role_id=$3", res.ID, e.GuildID, strings.Split(string(button.CustomID), "_")[1])
			}

			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Content: option.NewNullableString(""),
					Embeds: &[]discord.Embed{
						{
							Title:       "You can relax now 😅",
							Description: "We've sent your role selecter to the target channel!\n\n-# You can dismis this message safely",
							Color:       utils.Green,
						},
					},
					Components: &discord.ContainerComponents{},
				},
			})
		},
	}

	// Give / revoke the role
	MapButtonComponents["gib"] = ButtonComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ButtonInteraction) {

			target := strings.Split(string(data.CustomID), "_")[1]

			exists := false

			for _, role := range e.Member.RoleIDs {
				if role.String() == target {
					exists = true
				}
			}

			if exists {
				core.State.RemoveRole(e.GuildID, e.SenderID(), discord.RoleID(utils.MustSnowflakeEnv(target)), api.AuditLogReason(e.Member.User.Username+" updated their roles!"))
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "Done!",
								Description: "Role removed successfully",
								Color:       utils.Red,
							},
						},
					},
				})
				return
			}

			row := core.DB.QueryRow("SELECT password FROM interaction_roles WHERE guild_id=$1 AND role_id=$2", e.GuildID, target)

			var role InteractionRole
			row.Scan(&role.Password)

			if role.Password != "" {
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.ModalResponse,
					Data: &api.InteractionResponseData{
						CustomID: option.NewNullableString(string(data.CustomID)),
						Title:    option.NewNullableString("This role is protected!"),
						Components: &discord.ContainerComponents{
							&discord.ActionRowComponent{
								&discord.TextInputComponent{
									CustomID: "password",
									Style:    discord.TextInputShortStyle,
									Label:    "Password",
								},
							},
						},
					},
				})
				return
			}

			err := core.State.AddRole(e.GuildID, e.SenderID(), discord.RoleID(utils.MustSnowflakeEnv(target)), api.AddRoleData{AuditLogReason: api.AuditLogReason(e.Member.User.Username + " updated their roles!")})
			if err != nil {
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "thats awkward...",
								Description: "We cant give you that role right now...\n\nThere is a high chance that RoleCall wasnt configured properly, contact your guild's administrator with this error:\n```" + err.Error() + "```\nIf you ARE the server administrator, please reach out for support.",
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
							Title:       "Done!",
							Description: "Role applied successfully",
							Color:       utils.Green,
						},
					},
				},
			})
			core.DB.Exec("UPDATE interaction_roles SET uses=uses+1 WHERE guild_id=$1 AND role_id=$2", e.GuildID, target)
		},
	}
	MapModalComponents["gib"] = ModalComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.ModalInteraction) {
			target := strings.Split(string(data.CustomID), "_")[1]
			input := data.Components.Find("password").(*discord.TextInputComponent).Value

			row := core.DB.QueryRow("SELECT password FROM interaction_roles WHERE guild_id=$1 AND role_id=$2", e.GuildID, target)

			var entry InteractionRole
			row.Scan(&entry.Password)

			if !utils.VerifyPassword(input, entry.Password) {
				core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.MessageInteractionWithSource,
					Data: &api.InteractionResponseData{
						Flags: discord.EphemeralMessage,
						Embeds: &[]discord.Embed{
							{
								Title:       "Nice try!",
								Description: "The password you entered did not match!",
								Color:       utils.Red,
							},
						},
					},
				})
				return
			}

			core.State.AddRole(e.GuildID, e.SenderID(), discord.RoleID(utils.MustSnowflakeEnv(target)), api.AddRoleData{AuditLogReason: api.AuditLogReason(e.Member.User.Username + " got a password right and updated their roles!")})
			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.MessageInteractionWithSource,
				Data: &api.InteractionResponseData{
					Flags: discord.EphemeralMessage,
					Embeds: &[]discord.Embed{
						{
							Title:       "Done!",
							Description: "Role applied successfully",
							Color:       utils.Green,
						},
					},
				},
			})
			core.DB.Exec("UPDATE interaction_roles SET uses=uses+1 WHERE guild_id=$1 AND role_id=$2", e.GuildID, target)
		},
	}

	// Configuring autorole
	MapRoleComponents["autorole"] = RoleSelectComponent{
		Run: func(e *gateway.InteractionCreateEvent, data *discord.RoleSelectInteraction) {
			target := data.Values[0]

			core.DB.Exec("INSERT INTO autoroles (guild_id, role_id) VALUES ($1,$2) ON CONFLICT (guild_id) DO UPDATE SET role_id=EXCLUDED.role_id", e.GuildID, target)

			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &api.InteractionResponseData{
					Embeds: &[]discord.Embed{
						{
							Title:       "Configured",
							Description: "We'll give all new members the role <@&" + target.String() + ">",
							Color:       utils.Green,
						},
					},
					Components: &discord.ContainerComponents{},
				},
			})
		},
	}
}
