package bot

import (
	"fin-bot-miniapp/internal/logger"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) recoverMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered: %v", r)
				_ = c.Send("❌ Внутренняя ошибка сервера. Попробуйте позже.")
			}
		}()
		return next(c)
	}
}

func (h *Handler) errorMiddleware(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if err := next(c); err != nil {
			logger.Error("bot handler error: %v", err)
			return nil
		}
		return nil
	}
}
