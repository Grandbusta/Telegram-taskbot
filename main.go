package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println(".env loaded succesfully")

}

type Msg struct {
	FirstName   string
	ChatID      int64
	MessageID   int
	MessageText string
}

type CommandEntity struct {
	Command string
	Desc    string
}

var commands = []CommandEntity{
	{
		Command: "/start",
		Desc:    "Start",
	},
	{
		Command: "/add_task",
		Desc:    "Add a new task",
	},
	{
		Command: "/active_tasks",
		Desc:    "List active tasks",
	},
}

func initCommands(bot *tgbotapi.BotAPI) {
	tg_commands := make([]tgbotapi.BotCommand, 0, len(commands))
	for _, cmd := range commands {
		tg_commands = append(tg_commands, tgbotapi.BotCommand{Command: cmd.Command, Description: cmd.Desc})
	}
	_, err := bot.Request(tgbotapi.NewSetMyCommands(tg_commands...))
	if err != nil {
		panic(err)
	}
}

func getCommandsStrList() (list string) {
	str_commands := make([]string, 0)
	for _, cmd := range commands {
		if cmd.Command != "/start" {
			str_commands = append(str_commands, cmd.Command)
		}
	}
	list = strings.Join(str_commands[:], "\n")
	return list
}

func sendStartMsg(bot_instance *tgbotapi.BotAPI, msg_details *tgbotapi.Message) (err error) {
	msg := fmt.Sprintf(`
Hi %v, Welcome to your Assistant Bot.🤖

I help you manage your tasks, remind you of upcoming tasks/events that you set.😊🚀

Made with ❤️ in 🇳🇬

What do you wanna do?

%v
			`, msg_details.From.FirstName, getCommandsStrList())
	new_msg := tgbotapi.NewMessage(msg_details.Chat.ID, msg)
	if _, err := bot_instance.Send(new_msg); err != nil {
		return err
	}
	return nil

}
func addTaskMsg(bot_instance *tgbotapi.BotAPI, msg_details *tgbotapi.Message) (err error) {
	msg := fmt.Sprintf("Input a task to add")
	new_msg := tgbotapi.NewMessage(msg_details.Chat.ID, msg)
	new_msg.ReplyMarkup = tgbotapi.ForceReply{
		ForceReply:            true,
		InputFieldPlaceholder: msg,
		Selective:             false,
	}
	if _, err := bot_instance.Send(new_msg); err != nil {
		return err
	}
	return nil
}

var all_task = make([]string, 0)

func addNewTask(text string) {
	all_task = append(all_task, fmt.Sprintf("⭕ %v", text))
}

func handleMessage(bot_instance *tgbotapi.BotAPI, msg_details *tgbotapi.Message) (err error) {
	if msg_details.ReplyToMessage == nil {
		msg := fmt.Sprintf("Unrecognized\n\nCommands:\n%v", getCommandsStrList())
		new_msg := tgbotapi.NewMessage(msg_details.Chat.ID, msg)
		if _, err := bot_instance.Send(new_msg); err != nil {
			return err
		}
		return nil
	}
	if msg_details.ReplyToMessage.Text == "Input a task to add" {
		addNewTask(msg_details.Text)
		msg := fmt.Sprintf("Task \"%v\" added successfully✅\n\n%v", msg_details.Text, getCommandsStrList())
		new_msg := tgbotapi.NewMessage(msg_details.Chat.ID, msg)
		if _, err := bot_instance.Send(new_msg); err != nil {
			return err
		}
		err = listActiveTasks(bot_instance, msg_details)
		if err != nil {
			return err
		}
	}
	return nil
}

func listActiveTasks(bot_instance *tgbotapi.BotAPI, msg_details *tgbotapi.Message) (err error) {
	no_active_msg := fmt.Sprintf("No active task\n\n%v", getCommandsStrList())
	if len(all_task) == 0 {
		new_msg := tgbotapi.NewMessage(msg_details.Chat.ID, no_active_msg)
		if _, err := bot_instance.Send(new_msg); err != nil {
			return err
		}
		return nil
	}
	list := strings.Join(all_task[:], "\n\n")
	msg := fmt.Sprintf("Active tasks:\n\n%v\n\n%v", list, getCommandsStrList())
	new_msg := tgbotapi.NewMessage(msg_details.Chat.ID, msg)
	if _, err := bot_instance.Send(new_msg); err != nil {
		return err
	}
	return nil
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TG_API_KEY"))
	if err != nil {
		panic(err)
	}
	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)
	initCommands(bot)

	for update := range updates {

		if update.Message == nil {
			continue
		}

		fmt.Println("messagesas:", update.Message.ReplyToMessage)

		switch update.Message.Command() {
		case "":
			err = handleMessage(bot, update.Message)
		case "start":
			err = sendStartMsg(bot, update.Message)
		case "add_task":
			err = addTaskMsg(bot, update.Message)
		case "active_tasks":
			err = listActiveTasks(bot, update.Message)
		}
	}

}
