package services

import (
	"context"
	"log"
	"sync"

	"kingultron99.com/RoleCall/core"
	"kingultron99.com/RoleCall/middleware"
	"kingultron99.com/RoleCall/utils"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

type BroadcastResult struct {
	GuildID discord.GuildID
	Success bool
	Error   error
}

func BroadcastMessage(ctx context.Context, messages []BroadCastMessage) []BroadcastResult {
	guilds, err := core.State.Guilds()
	if err != nil {
		log.Printf("failed to fetch guilds: %v", err)
		return nil
	}

	results := make([]BroadcastResult, 0, len(guilds))

	var wg sync.WaitGroup
	var mu sync.Mutex

	// limit concurrency so you don't smash rate limits
	sem := make(chan struct{}, 5)

	for _, g := range guilds {
		wg.Add(1)

		go func(g discord.Guild) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			fullGuild, err := core.State.Guild(g.ID)
			if err != nil {
				appendResult(&mu, &results, BroadcastResult{
					GuildID: g.ID,
					Success: false,
					Error:   err,
				})
				return
			}

			var (
				systemEnabled, updatesEnabled bool
				location                      *discord.ChannelID = nil
			)
			err = core.DB.QueryRow("SELECT system, updates, location FROM broadcast_preferences WHERE guild_id=$1", g.ID).Scan(&systemEnabled, &updatesEnabled, &location)
			if err != nil {
				log.Printf("failed to fetch broadcast preferences for guild %s: %v", g.ID, err)
				appendResult(&mu, &results, BroadcastResult{
					GuildID: g.ID,
					Success: false,
					Error:   err,
				})
				return
			}

			var (
				messageList []BroadCastMessage
				embeds      []discord.Embed
				footerText  string = "-# If you have any questions or concerns, please join the support server.\n\n-# If you are a server owner and would like to stop receiving these messages, you can disable them with the `/configure broadcasts` command.\n\n-# If you have specified a custom channel and are still not receiving messages there, make sure RoleCall has permission to view and send messages in that channel."
			)
			for _, message := range messages {
				// If the guild has opted out of a message type, skip adding that message to the list of messages to send.
				if message.Type == SystemNoticeMessage && !systemEnabled || message.Type == NewVersionMessage && !updatesEnabled {
					continue
				}
				messageList = append(messageList, message)
			}
			// If the guild has opted out of all messages, skip the guild.
			if len(messageList) == 0 {
				return
			}

			claims := middleware.GetClaims(ctx)

			author, err := core.State.User(discord.UserID(utils.MustSnowflakeEnv(claims.Subject)))
			if err != nil {
				log.Println("Failed to get Author")
			}

			for _, message := range messageList {
				embeds = append(embeds, discord.Embed{
					Title: BroadCastMessageType[message.Type].Title,
					Color: BroadCastMessageType[message.Type].Color,
					// Author: &discord.EmbedAuthor{
					// },
					Description: message.Content,
					Footer: &discord.EmbedFooter{
						Text: author.Username + " authored this broadcast!",
						Icon: author.AvatarURL(),
					},
					Timestamp: discord.NowTimestamp(),
				})
			}

			// Add short info/instruction footer embed
			embeds = append(embeds, discord.Embed{
				Description: footerText,
			})

			// Determine the channel to send the message in based on the guild's preferences and available channels.
			// Fall back to system channel if no custom location is set, then public updates channel, then fail if neither are set.
			//
			// This order is somewhat arbitrary, but it prioritizes the channel that the server owner explicitly chose, then the system
			// channel which is often used for important announcements, then the public updates channel which is less commonly used but
			// still a reasonable default.
			//
			// If neither of those channels are set, we skip the guild and log an error, rather than sending the message to a random
			// channel which could be disruptive and annoying for users.

			if location != nil && location.IsValid() {
				_, err = core.State.SendMessageComplex(*location, api.SendMessageData{
					Embeds: embeds,
				})
				if err != nil {
					log.Printf("failed to send to configured channel for guild %v. Err: %v", g.ID, err)
				} else {
					appendResult(&mu, &results, BroadcastResult{
						GuildID: g.ID,
						Success: true,
						Error:   nil,
					})
					return
				}
			}

			_, err = core.State.SendMessageComplex(fullGuild.PublicUpdatesChannelID, api.SendMessageData{
				Embeds: embeds,
			})

			// If sending to the public channel fails for whatever reason, try sending to the system channel before giving up and erroring out.
			if err != nil {
				err = nil // reset err so we don't log it twice if the system channel also fails
				_, err = core.State.SendMessageComplex(fullGuild.SystemChannelID, api.SendMessageData{
					Embeds: embeds,
				})

				if err != nil {
					log.Printf("failed to send message to guild %s: %v", g.ID, err)
					appendResult(&mu, &results, BroadcastResult{
						GuildID: g.ID,
						Success: false,
						Error:   errNoBroadcastChannel,
					})
					return
				} else {
					appendResult(&mu, &results, BroadcastResult{
						GuildID: g.ID,
						Success: true,
						Error:   nil,
					})
					return
				}
			} else {
				appendResult(&mu, &results, BroadcastResult{
					GuildID: g.ID,
					Success: true,
					Error:   nil,
				})
				return
			}

		}(g)
	}

	wg.Wait()

	return results
}
