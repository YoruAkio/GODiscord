package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Command represents a bot command
type Command struct {
	Name        string
	Description string
	Handler     func(s *discordgo.Session, m *discordgo.MessageCreate)
}

func main() {
	startTime := time.Now()

	// Load the .env file
	if err := godotenv.Load(); err != nil {
		logError("Error loading .env file:", err)
		return
	}

	// Get your bot token from the Discord Developer Portal
	token := os.Getenv("TOKEN")
	if token == "" {
		logError("Please set the TOKEN environment variable.", nil)
		return
	}

	// Create a new Discord session
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		logError("Error creating Discord session:", err)
		return
	}

	// Define the list of commands
	commands := make([]*Command, 3)

	commands[0] = &Command{
		Name:        "ping",
		Description: "Replies with 'Pong!' and shows response time and client WebSocket ping.",
		Handler:     pingCommand,
	}

	commands[1] = &Command{
		Name:        "echo",
		Description: "Repeats back the message sent after the command.",
		Handler:     echoCommand,
	}

	commands[2] = &Command{
		Name:        "help",
		Description: "Shows the list of available commands.",
		Handler:     func(s *discordgo.Session, m *discordgo.MessageCreate) { helpCommand(s, m, commands) },
	}

	// Register message handler function
	discord.AddHandler(messageHandler(commands))

	defer discord.Close()

	// Set the bot's presence
	discord.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: "chilling with Gopher",
				Type: discordgo.ActivityTypeGame,
			},
		},
		Status: "online",
	})

	// Open a websocket connection to Discord
	if err = discord.Open(); err != nil {
		logError("Error opening connection:", err)
		return
	}

	// Calculate the startup time
	startupTime := time.Since(startTime).Round(time.Millisecond)

	logInfo("Bot is now running. Startup time: " + startupTime.String())
	logInfo("Press CTRL-C to exit.")

	// Wait here until CTRL-C or other term signal is received
	sc := make(chan os.Signal, 1)
	<-sc
}

func logInfo(message string) {
	fmt.Printf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}

func logError(message string, err error) {
	fmt.Printf("[%s] %s: %v\n", time.Now().Format("2006-01-02 15:04:05"), message, err)
}

func logCommandExecution(s *discordgo.Session, m *discordgo.MessageCreate) {
	server, _ := s.Guild(m.GuildID)
	channel, _ := s.Channel(m.ChannelID)
	fmt.Printf("[%s] %s#%s (%s) in #%s: %s\n", time.Now().Format("2006-01-02 15:04:05"), m.Author.Username, m.Author.Discriminator, server.Name, channel.Name, m.Content)
}

func messageHandler(commands []*Command) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore messages from the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		// Check if the message starts with the command prefix
		if !strings.HasPrefix(m.Content, "!") {
			return
		}

		// Extract the command name
		fields := strings.Fields(m.Content)
		if len(fields) == 0 {
			return
		}
		cmdName := fields[0][1:]

		// Find the matching command handler
		for _, cmd := range commands {
			if cmd.Name == cmdName {
				cmd.Handler(s, m)
				return
			}
		}

		// Command not found
		s.ChannelMessageSend(m.ChannelID, "Invalid command. Try !help for a list of commands.")
	}
}

func pingCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	logCommandExecution(s, m)
	start := time.Now()
	msg, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if err != nil {
		logError("Error sending message:", err)
		return
	}
	elapsed := time.Since(start)
	wsLatency := s.HeartbeatLatency().Round(time.Millisecond)
	content := fmt.Sprintf("Pong! Response Time: %s | WebSocket Ping: %s", elapsed, wsLatency)
	s.ChannelMessageEdit(msg.ChannelID, msg.ID, content)
}

func echoCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	logCommandExecution(s, m)
	content := strings.TrimPrefix(m.Content, "!echo ")
	if content != "" {
		s.ChannelMessageSend(m.ChannelID, content)
	} else {
		s.ChannelMessageSend(m.ChannelID, "Please provide something to echo!")
	}
}

func helpCommand(s *discordgo.Session, m *discordgo.MessageCreate, commands []*Command) {
	logCommandExecution(s, m)
	var helpMessage strings.Builder
	helpMessage.WriteString("Available commands:\n")
	for _, cmd := range commands {
		helpMessage.WriteString(fmt.Sprintf("- !%s: %s\n", cmd.Name, cmd.Description))
	}
	s.ChannelMessageSend(m.ChannelID, helpMessage.String())
}
