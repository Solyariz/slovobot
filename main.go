package main
import (

	"log"
	"io/ioutil"
	"strings"
	"fmt"
	//"unicode/utf8"
	"time"
	"math/rand"

	"github.com/bot-api/telegram"
	"github.com/bot-api/telegram/telebot"
	"golang.org/x/net/context"
	"os"
	"slovobot/util"
)

type words struct {
	used []string
	dictionary map[rune][]string
	lastWord string
}

const (
	START = "start"
	HELP  = "help"
	LIST  = "list"
	RESET  = "reset"
)


func main() {
	token := os.Getenv("BOT_TOKEN")
	commands := make([]string, 4)
	commands[0] = START
	commands[1] = HELP
	commands[2] = LIST
	commands[2] = RESET
	// connect to bot with token
	api := telegram.New(token)
	api.Debug(true)
	bot := telebot.NewWithAPI(api)

	// init
	bot.Use(telebot.Recover()) // recover if handler panics
	netCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dict := load()
	fmt.Println("LENGTH: ", len(dict))

	var game words

	bot.Use(telebot.Commands(map[string]telebot.Commander{
		START: telebot.CommandFunc(
			func(ctx context.Context, arg string) error {

				api := telebot.GetAPI(ctx)
				update := telebot.GetUpdate(ctx)
				_, err := api.SendMessage(ctx,
					telegram.NewMessagef(update.Chat().ID,
						"received START with arg %s", arg,
					))
				game = words{make([]string, 0), dict, ""}
				return err
			}),
		RESET: telebot.CommandFunc(
			func(ctx context.Context, arg string) error {

				api := telebot.GetAPI(ctx)
				update := telebot.GetUpdate(ctx)
				_, err := api.SendMessage(ctx,
					telegram.NewMessagef(update.Chat().ID,
						"received reset with arg %s", arg,
					))
				game = words{make([]string, 0), dict, ""}
				return err
			}),
		HELP: telebot.CommandFunc(
			func(ctx context.Context, arg string) error {

				api := telebot.GetAPI(ctx)
				update := telebot.GetUpdate(ctx)
				//command, arg := update.Message.Command()
				
				text := fmt.Sprintf("%s", commands)
				fmt.Println(text)
				_, err := api.SendMessage(ctx,
					telegram.NewMessage(update.Chat().ID,
						text))
				return err
			}),
		LIST: telebot.CommandFunc(
			func(ctx context.Context, arg string) error {

				api := telebot.GetAPI(ctx)
				update := telebot.GetUpdate(ctx)
				//command, arg := update.Message.Command()
				limit := 100
				wordList := game.dictionary[firstLetter(arg)]
				filtered := make([]string, 0)
				for _, el := range wordList {
					if strings.HasPrefix(el, arg) {
						filtered = append(filtered, el)
					}
				}
				if len(filtered) < limit {
					limit = len(filtered)
				}
				text := fmt.Sprintf("%s", filtered[0:limit])
				fmt.Println(text)
				_, err := api.SendMessage(ctx,
					telegram.NewMessage(update.Chat().ID,
						text))
				return err
			}),
		"": telebot.CommandFunc(
			func(ctx context.Context, arg string) error {

				api := telebot.GetAPI(ctx)
				update := telebot.GetUpdate(ctx)
				command, arg := update.Message.Command()
				_, err := api.SendMessage(ctx,
					telegram.NewMessagef(update.Chat().ID,
						"received unrecognized"+
							" command %s with arg %s",
						command, arg,
					))
				return err
			}),
	}))

	//bot.Handle(telebot.StringHandler(buildReply(game, )))
	bot.HandleFunc(func (ctx context.Context) error {
		update := telebot.GetUpdate(ctx) // take update from context
		if update.Message == nil {
			return nil
		}
		var msg string
		if len(game.dictionary) == 0 {
			msg = "Game is not started yet. Try /help"
		} else {
			msg = buildReply(game, update.Message.Text)
		}
		api := telebot.GetAPI(ctx) // take api from context
		_, err := api.Send(ctx, telegram.NewMessage(update.Message.Chat.ID, msg))
		return err
	})

	err := bot.Serve(netCtx)
	if err != nil {
		log.Fatal(err)
	}
}
func buildReply(game words, Text string) string {
	var reply string
	if !game.startsWithRightLetter(Text) {
		reply = "Word must start with last letter from previous word"
	} else if !game.isCorrect(Text) {
		reply = "There's no such word! Try another one."
	} else if game.isUsed(Text) {
		reply = "This word has already been used! Try another one."
	} else {
		game.use(Text)
		game.lastWord = Text
		reply = game.chooseReply(Text)
		if reply == "" {
			reply = "Sorry, I'm stupid. Try another word."
		} else {
			game.lastWord = reply
		}
	}
	fmt.Println("reply: ", reply)
	return reply
}

func load() (map[rune][]string) {
	file, err := ioutil.ReadFile("./dict/words_rus")
	if err != nil {
		panic(err)
	}
	wordMap := make(map[rune][]string)
	words := strings.Fields(string(file))
	fmt.Println(len(words))
	for _, el := range words {
		runeWord := util.ToRunes(el)
		firstLetter := runeWord[0]
		wordList, ok := wordMap[firstLetter]
		if ok {
			wordList = append(wordList, el)
			wordMap[firstLetter] = wordList
		} else {
			wordList = make([]string, 1)
			wordList[0] = el
			//wordList = append(wordList, el)
			wordMap[firstLetter] = wordList
		}
	}
	return wordMap
}



//func chooseReply(msg string, game words) (string) {
//	lastLetter := lastLetter(msg)
//	fmt.Println(lastLetter)
//	//fmt.Println("list len: ", len(dict[lastLetter]))
//	length := len(game.dictionary[lastLetter])
//	s1 := rand.NewSource(time.Now().UnixNano())
//	r1 := rand.New(s1)
//	candidate := game.dictionary[lastLetter][r1.Intn(length-1)]
//	ok := isUsed(game, candidate)
//	for ok == true {
//		candidate = game.dictionary[lastLetter][r1.Intn(length-1)]
//	}
//	game.used = append(game.used, candidate)
//	fmt.Println(game.used)
//	return candidate
//}

func (game * words) chooseReply(msg string) string {
	rWord := util.ToRunes(msg)
	i := len(rWord) - 1
	lastLetter := rWord[i]
	//fmt.Println(lastLetter)
	//fmt.Println("list len: ", len(dict[lastLetter]))
	wordList := game.dictionary[lastLetter]

	length := len(wordList)

	for length == 0 && i >= 0{
		i--
		lastLetter = rWord[i]
		wordList = game.dictionary[lastLetter]
		length = len(wordList)
	}

	if length == 0 {
		return ""
	}

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	candidate := wordList[r1.Intn(length-1)]
	ok := game.isUsed(candidate)
	for ok == true {
		candidate = wordList[r1.Intn(length-1)]
	}
	game.used = append(game.used, candidate)
	fmt.Println(game.used)
	return candidate
}

func (game * words) isCorrect(word string) bool {
	fLetter := firstLetter(word)
	//fmt.Println(game.dictionary[fLetter])
	return contains(word, game.dictionary[fLetter])
}

func (game * words) startsWithRightLetter(word string) bool {
	if len(game.lastWord) == 0 {
		return true
	}
	rWord := util.ToRunes(game.lastWord)
	i := len(rWord)-1
	last := rWord[i]
	for len(game.dictionary[last]) == 0 && i >= 0 {
		i--
		last = rWord[i]
	}
	first := firstLetter(word)
	if last == -1{
		return true;
	} else if first == -1 {
		panic("This can't happen")
	} else {
		return first == last
	}
}

func lastLetter(msg string) rune {
	if msg == "" {
		return -1
	} else {
		runes := util.ToRunes(msg)
		//fmt.Println(runes)
		lastLetter := runes[len(runes)-1]
		return lastLetter
	}
}

func firstLetter(msg string) rune {
	if msg == "" {
		return -1
	} else {
		runes := util.ToRunes(msg)
		//fmt.Println(runes)
		lastLetter := runes[0]
		return lastLetter
	}
}

func (game * words) isUsed(word string) bool {
	return contains(word, game.used)
}

func contains(word string, words []string) bool {
	for _, w := range words{
		if w==word {
			return true
		}
	}
	return false
}

func (game * words) use(word string) {
	game.used = append(game.used, word)
}

