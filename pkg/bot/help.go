package bot

import tele "gopkg.in/telebot.v3"

func HandleHelpCommand(c tele.Context) error {
	return c.Reply(`/help - show this help message
/gpt <message> - start a new chat
/set <message> - set the system content
/tc <message> - translate message to Chinese
`)
}
