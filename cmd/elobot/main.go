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

	session, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
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
