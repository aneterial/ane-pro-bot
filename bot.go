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
var ms *MessageStash
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
		logging("he mod enabled")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "he mod enabled",
			},
		})
	},
	"he_disable": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		heMod = false
		logging("he mod disabled")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "he mod disabled",
			},
		})
	},
}

func init() {
	ms = new(MessageStash)
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

		return
	}

	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!!!")

		return
	}

	if heMod {
		if ms.Empty() {
			ms.Fill(m.Author.ID, m.ChannelID, m.Content)

			return
		} else if ms.CheckOverflow(m.ChannelID, m.Content) {
			go sendMessage(s, m.ChannelID, m.Content)
		}

		ms.Flush()

		return
	}

	if heMod && m.Content == "👋" {
		s.ChannelMessageSend(m.ChannelID, "👋")
	}
	if strings.Contains(m.Content, "what is") {
		s.ChannelMessageSend(m.ChannelID, "У тебя прав нет, фрикич<:he:856194932770471936>")
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

func sendMessage(s *discordgo.Session, ChannelID string, message string) {
	time.Sleep(1 * time.Second)
	s.ChannelMessageSend(ChannelID, message)
}

type MessageStash struct {
	authorID  string
	channelID string
	message   string
}

func (ms *MessageStash) Empty() bool {
	return ms.channelID == ""
}

func (ms *MessageStash) CheckOverflow(channelID string, message string) bool {
	return ms.channelID == channelID && ms.message == message
}

func (ms *MessageStash) Fill(authorID string, channelID string, message string) {
	ms.authorID = authorID
	ms.channelID = channelID
	ms.message = message
}

func (ms *MessageStash) Flush() {
	ms.authorID = ""
	ms.channelID = ""
	ms.message = ""
}
