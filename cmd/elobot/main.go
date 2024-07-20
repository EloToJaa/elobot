package main

import (
	"os"
	"os/signal"

	cmd "github.com/EloToJaa/elobot/internal/commands"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var session *discordgo.Session

var (
	BotToken       string
	GuildId        string
	RemoveCommands = true
)

var (
	commands        = cmd.Commands
	commandHandlers = cmd.CommandHandlers
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})

	log.SetOutput(os.Stdout)

	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(false) // add filename & function to logs

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	BotToken = os.Getenv("DISCORD_TOKEN")
	GuildId = os.Getenv("DISCORD_GUILD_ID")
}

func main() {
	log.Info("Start program")

	session.AddHandler(func(sess *discordgo.Session, r *discordgo.Ready) {
		log.Info("Bot is up")
	})
	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err := session.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer session.Close()

	createdCommands, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, GuildId, commands)
	if err != nil {
		log.Fatalf("Cannot register commands: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Info("Gracefully shutting down")

	if RemoveCommands {
		for _, cmd := range createdCommands {
			err := session.ApplicationCommandDelete(session.State.User.ID, GuildId, cmd.ID)
			if err != nil {
				log.Fatalf("Cannot delete %q command: %v", cmd.Name, err)
			}
		}
	}
}

func createDiscordSession(token string) *discordgo.Session {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session", err)
	}

	session.ShouldReconnectOnError = true // Not sure if this is needed
	return session
}

func startBotHandlers(session *discordgo.Session) map[string]map[string]botsdef.Discord {
	// bots := make(map[string]map[string]botsdef.Discord)

	// guildIDs := []string{"123"}

	// for _, id := range guildIDs {
	// 	bots[id] = make(map[string]botsdef.Discord)

	// 	prefix := "!"

	// 	for _, module := range botsdef.Modules {
	// 		botInstance := botsdef.CreateBotInstance(session, module)
	// 		if botInstance != nil {
	// 			bots[id][module] = botInstance
	// 			botInstance.Start(id, prefix)
	// 		}
	// 	}
	// }

	// guildManager := manager.NewGuildManager(session, bots)
	// guildManager.Start()

	return bots
}

func handleDiscordSession(discordSession *discordgo.Session) {
	if err := discordSession.Open(); err != nil {
		log.Fatal("Error opening Discord session", err)
		os.Exit(1)
	}
	defer discordSession.Close()
}
