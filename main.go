package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"kingultron99.com/RoleCall/core"
	"kingultron99.com/RoleCall/events"
	"kingultron99.com/RoleCall/utils"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session/shard"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/joho/godotenv"
)

func main() {

	devPtr := flag.Bool("dev", false, "Should RoleCall run in dev mode?")

	if *devPtr {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	var token = os.Getenv("TOKEN")
	if token == "" {
		log.Panic("NO TOKEN PROVIDED!")
	}

	core.StartDB()

	newShard := state.NewShardFunc(func(m *shard.Manager, s *state.State) {
		// Add the needed Gateway intents.
		s.AddIntents(gateway.IntentGuilds)
		s.AddIntents(gateway.IntentGuildMessages)
		s.AddIntents(gateway.IntentGuildMembers)

		core.State = s
	})
	m, err := shard.NewManager("Bot "+token, newShard)
	if err != nil {
		log.Fatal("failed to create shard manager:", err)
	}

	if err := m.Open(context.Background()); err != nil {
		log.Fatal("failed to connect shards:", err)
	}

	var shardNum int

	m.ForEach(func(s shard.Shard) {
		state := s.(*state.State)

		u, err := state.Me()
		if err != nil {
			log.Fatal("failed to get myself:", err)
		}

		getGuilds, _ := state.Guilds()

		log.Printf("Shard %d/%d started as %s in %v guilds", shardNum, m.NumShards()-1, u.Tag(), len(getGuilds))

		shardNum++

		for _, guild := range getGuilds {
			go func() {
				events.RegisterCommands(discord.AppID(utils.MustSnowflakeEnv(os.Getenv("APP_ID"))), guild.ID)
			}()
		}
	})

	events.AddJoinHandler()
	events.CommandRouter()

	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Print("Preparing to shut down the server...")
		log.Print("closing shard manager...")
		m.Close()
		log.Print("disconnecting DB...")
		core.DB.Close()
		log.Print("bye!")
		os.Exit(0)
	}()
	// Block forever.
	select {}

}
