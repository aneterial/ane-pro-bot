package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const LOG_FILE string = "/var/log/ane_bot/app.log"

var heMod bool = true
var Token string
var dmPermission bool = false
var dg *discordgo.Session
var defaultMemberPermissions int64 = discordgo.PermissionManageServer
var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "time",
		Description:              "get time with nano sec",
		DefaultPermission:        &dmPermission,
		DefaultMemberPermissions: &defaultMemberPermissions,
	},
	{
		Name:                     "he_enable",
		Description:              "enable he mod",
		DefaultPermission:        &dmPermission,
		DefaultMemberPermissions: &defaultMemberPermissions,
	},
	{
		Name:                     "he_disable",
		Description:              "dicable he mod",
		DefaultPermission:        &dmPermission,
		DefaultMemberPermissions: &defaultMemberPermissions,
	},
}
var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"time": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "current time is: " + time.Now().Format("2006-01-02 15:04:05.999999999"),
			},
		})
	},
	"he_enable": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		heMod = true
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "he mod enabled",
			},
		})
	},
	"he_disable": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		heMod = false
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "he mod disabled",
			},
		})
	},
}

func init() {
	log.SetOutput(os.Stdout)
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	var err error
	dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

func main() {
	var err error

	dg.AddHandler(messageCreate)
	dg.AddHandler(baseHandler)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	defer dg.Close()

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		log.Println("add cmd", v.Name)
		registeredCommands[i] = cmd
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	log.Println("Bot exit with code", <-sc)
	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
		log.Println("remove cmd ", v.Name)
	}

	log.Println("Bot close")
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!!!")
	}

	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!!!")
	}

	if heMod && (m.Content == ":he~1:" || m.Content == ":he:") {
		s.ChannelMessageSend(m.ChannelID, ":he:")
	}

	if heMod && m.Content == "👋" {
		s.ChannelMessageSend(m.ChannelID, "👋")
	}
	if strings.Contains(m.Content, "what is") {
		s.ChannelMessageSend(m.ChannelID, "logging...")
		logging(m.Content)

	}
}

func baseHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
		h(s, i)
	}
}

func logging(s string) {
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println(s)
}
