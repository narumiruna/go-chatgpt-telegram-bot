package bot

import tele "gopkg.in/telebot.v3"

func onHelp(c tele.Context) error {
	return c.Reply(`/help - show this help message
/gpt <message> - start a new chat
/set <message> - set the system content
/temperature <temperature> - set the temperature`)
}
