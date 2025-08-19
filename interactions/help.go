package interactions

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"kingultron99.com/RoleCall/core"
	"kingultron99.com/RoleCall/utils"
)

func init() {
	MapCommands["help"] = Command{
		CreateCommandData: &api.CreateCommandData{
			Name:                     "help",
			Description:              "Information about RoleCall, and how to use it",
			DefaultMemberPermissions: discord.NewPermissions(discord.PermissionAdministrator),
			NoDMPermission:           true,
			Type:                     discord.ChatInputCommand,
		},
		Run: func(e *gateway.InteractionCreateEvent, data *discord.CommandInteraction) {
			core.State.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
				Type: api.MessageInteractionWithSource,
				Data: &api.InteractionResponseData{
					Flags: discord.EphemeralMessage,
					Embeds: &[]discord.Embed{
						{
							Description: `# Hi there! :wave:
### First off, Thanks for choosing RoleCall!
							
If you've somehow found yourself with RoleCall in your guild, and you're not quite sure what it is, here's a quick summary of its features:`,
							Color: utils.Purple,
						},
						{
							Title:       "Role providers",
							Description: "Effortlessly dispense out roles that your members can assign themselves with.",
							Fields: []discord.EmbedField{
								{
									Name:  "Password protected :closed_lock_with_key:",
									Value: "Sometimes, it'd be nice to restrict who has access to what roles when adding them to a provider... well now you can!",
								},
								{
									Name:  "Usage:",
									Value: "`/configure provider`",
								},
							},
							Color: utils.Purple,
						},
						{
							Title:       "Autoroles",
							Description: "AutoMagically:tm: dish out a \"default\" role to your guilds newest members!",
							Fields: []discord.EmbedField{
								{
									Name:  "Supports member screening",
									Value: "Whether you're a large community server equiped with discords member screening, or a small group of friends, RoleCall will make sure your newest members get your configured default role",
								},
								{
									Name:  "Usage:",
									Value: "`/configure autorole`",
								},
							},
							Color: utils.Purple,
						},
						{
							Title:       "Your roles. Your rules.",
							Description: "Passwords for roles are encrypted with Bcrypt, we collect as little information as possible and you will ALWAYS have the option to delete every configuration related to your guild",
							Fields: []discord.EmbedField{
								{
									Name:  "Usage:",
									Value: "`/delete <everything|autorole|provider|leftovers>`",
								},
							},
							Color: utils.Purple,
						},
					},
				},
			})
		},
	}
}
