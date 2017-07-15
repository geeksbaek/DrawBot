package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type credentials struct {
	Token string
}

// Errors
var (
	errNotEnoughArgs              = errors.New("The argument is not enough")
	errCouldNotFindUserVoiceState = errors.New("Could not find user's voice state")
)

// Constants
const (
	maxDiceNum = 100

	diceStringFormat = "<@%v>님이 주사위를 굴려서 **%v**이 나왔습니다. (1 ~ %v)"
)

// Variables used for command line parameters
var (
	Token string
	BotID string
)

func main() {
	// Get the Token from ./credentials.json
	file, err := ioutil.ReadFile("./credentials.json")
	if err != nil {
		log.Fatal(err)
	}

	credentials := &credentials{}
	json.Unmarshal(file, &credentials)

	Token = credentials.Token

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// Store the account ID for later use.
	BotID = u.ID

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(dice)
	dg.AddHandler(lots)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}

// 주사위 굴리기
func dice(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !(strings.HasPrefix(m.Content, "/dice") || strings.HasPrefix(m.Content, "/주사위")) {
		return
	}

	rand.Seed(int64(time.Now().Nanosecond()))
	randNum := rand.Intn(maxDiceNum - 1)
	randNum++ // To start from 1

	message := fmt.Sprintf(diceStringFormat, m.Author.ID, randNum, maxDiceNum)
	s.ChannelMessageDelete(m.ChannelID, m.Message.ID)
	s.ChannelMessageSend(m.ChannelID, message)
}

// 제비뽑기
func lots(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !(strings.HasPrefix(m.Content, "/lots") || strings.HasPrefix(m.Content, "/제비뽑기")) {
		return
	}

	vs, err := findUserVoiceState(s, m.Author.ID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(vs)
}

func mustGetArgs(content string) []string {
	args, err := getArgs(strings.Fields(content))
	if err != nil {
		log.Fatal(err)
	}
	return args
}

func getArgs(content []string) ([]string, error) {
	if len(content) < 2 {
		return nil, errNotEnoughArgs
	}
	return content[1:], nil
}

func findUserVoiceState(session *discordgo.Session, userid string) (*discordgo.VoiceState, error) {
	for _, guild := range session.State.Guilds {
		for _, vs := range guild.VoiceStates {
			if vs.UserID == userid {
				return vs, nil
			}
		}
	}
	return nil, errCouldNotFindUserVoiceState
}
