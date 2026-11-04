package bot

import (
	"fmt"
	"strings"

	"fin-bot-miniapp/internal/logger"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) handleStart(c tele.Context) error {
	return c.Send("👋 Добро пожаловать в Finance Bot!\n\nИспользуйте /add, чтобы записать расход или доход.\nВсе команды — /help.")
}

func (h *Handler) handleAdd(c tele.Context) error {
	if h.isGroup(c) {
		return c.Send("Чтобы добавлять записи, напишите мне в личные сообщения: @" + h.bot.Me.Username)
	}
	markup := &tele.ReplyMarkup{}
	btn := markup.WebApp("📝 Добавить запись", &tele.WebApp{URL: h.miniAppURL + "/app/add.html"})
	markup.Inline(markup.Row(btn))
	return c.Send("Нажмите кнопку, чтобы открыть форму:", markup)
}

func (h *Handler) handleStats(c tele.Context) error {
	if h.isGroup(c) {
		return c.Send("Чтобы посмотреть статистику, напишите мне в личные сообщения: @" + h.bot.Me.Username)
	}
	markup := &tele.ReplyMarkup{}
	btn := markup.WebApp("📊 Статистика", &tele.WebApp{URL: h.miniAppURL + "/app/stats.html"})
	markup.Inline(markup.Row(btn))
	return c.Send("Нажмите кнопку, чтобы посмотреть статистику:", markup)
}

func (h *Handler) handleExport(c tele.Context) error {
	if h.isGroup(c) {
		return c.Send("Чтобы выгрузить записи, напишите мне в личные сообщения: @" + h.bot.Me.Username)
	}
	markup := &tele.ReplyMarkup{}
	btn := markup.WebApp("📋 Выгрузить в Excel", &tele.WebApp{URL: h.miniAppURL + "/app/export.html"})
	markup.Inline(markup.Row(btn))
	return c.Send("Нажмите кнопку, чтобы выгрузить записи:", markup)
}

func (h *Handler) handleHistory(c tele.Context) error {
	if h.isGroup(c) {
		return c.Send("Чтобы посмотреть историю, напишите мне в личные сообщения: @" + h.bot.Me.Username)
	}
	markup := &tele.ReplyMarkup{}
	btn := markup.WebApp("📜 История", &tele.WebApp{URL: h.miniAppURL + "/app/history.html"})
	markup.Inline(markup.Row(btn))
	return c.Send("Нажмите кнопку, чтобы просмотреть записи:", markup)
}

func (h *Handler) handleUndo(c tele.Context) error {
	if h.isGroup(c) {
		return c.Send("Чтобы отменить запись, напишите мне в личные сообщения: @" + h.bot.Me.Username)
	}
	markup := &tele.ReplyMarkup{}
	btn := markup.WebApp("↩️ Отменить последнюю запись", &tele.WebApp{URL: h.miniAppURL + "/app/undo.html"})
	markup.Inline(markup.Row(btn))
	return c.Send("Нажмите кнопку, чтобы проверить и отменить:", markup)
}

func (h *Handler) handleRecurring(c tele.Context) error {
	if h.isGroup(c) {
		return c.Send("Чтобы управлять регулярными платежами, напишите мне в личные сообщения: @" + h.bot.Me.Username)
	}
	markup := &tele.ReplyMarkup{}
	btn := markup.WebApp("🔁 Регулярные платежи", &tele.WebApp{URL: h.miniAppURL + "/app/recurring.html"})
	markup.Inline(markup.Row(btn))
	return c.Send("Нажмите кнопку, чтобы управлять регулярными платежами:", markup)
}

func (h *Handler) handleAddCat(c tele.Context) error {
	args := c.Args()
	if len(args) < 3 {
		return c.Send("Использование: /addcat <код> <эмодзи> <название>\nПример: /addcat gym 🏋️ Спортзал")
	}
	code := args[0]
	emoji := args[1]
	name := strings.Join(args[2:], " ")

	userID := c.Sender().ID
	if err := h.categoryService.AddCategory(userID, code, name, emoji, 99); err != nil {
		logger.Error("AddCategory user=%d: %v", userID, err)
		return c.Send(fmt.Sprintf("❌ Не удалось добавить категорию: %v", err))
	}
	return c.Send(fmt.Sprintf("✅ Категория добавлена: %s %s", emoji, name))
}

func (h *Handler) handleAddLedger(c tele.Context) error {
	args := c.Args()
	if len(args) < 3 {
		return c.Send("Использование: /addledger <код> <эмодзи> <название>\nПример: /addledger savings 💰 Накопления")
	}
	code := args[0]
	emoji := args[1]
	name := strings.Join(args[2:], " ")

	userID := c.Sender().ID
	if err := h.ledgerService.AddLedger(userID, code, name, emoji, 99); err != nil {
		logger.Error("AddLedger user=%d: %v", userID, err)
		return c.Send(fmt.Sprintf("❌ Не удалось добавить счёт: %v", err))
	}
	return c.Send(fmt.Sprintf("✅ Счёт добавлен: %s %s", emoji, name))
}

func (h *Handler) handleHelp(c tele.Context) error {
	return c.Send(`Finance Bot — команды

/add        Записать расход или доход
/recurring  Регулярные платежи
/stats      Статистика
/export     Выгрузить записи в Excel
/undo       Отменить последнюю запись
/history    Просмотр и правка записей
/addcat     Добавить свою категорию
/addledger  Добавить свой счёт
/help       Показать это сообщение`)
}
