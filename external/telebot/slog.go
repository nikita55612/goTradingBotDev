package telebot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type BotSlogHandler struct {
	bot *Bot
	slog.Handler
}

func NewBotSlogHandler(apiKey string, writeChatID string, slogOpts *slog.HandlerOptions, botOpts ...Option) *BotSlogHandler {
	botOpts = append(botOpts, WithWriteChatID(writeChatID))
	return &BotSlogHandler{
		bot:     NewBot(apiKey, botOpts...),
		Handler: slog.NewTextHandler(os.Stdin, slogOpts),
	}
}

func NewBotSlogHandlerFromEnv(writeChatID string, slogOpts *slog.HandlerOptions, botOpts ...Option) *BotSlogHandler {
	botOpts = append(botOpts, WithWriteChatID(writeChatID))
	return &BotSlogHandler{
		bot:     NewBotFromEnv(botOpts...),
		Handler: slog.NewTextHandler(os.Stdin, slogOpts),
	}
}

func (h *BotSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *BotSlogHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *BotSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	switch r.Level {
	case slog.LevelDebug:
		level = "⚪️ " + level
	case slog.LevelInfo:
		level = "🟢 " + level
	case slog.LevelWarn:
		level = "🟡 " + level
	case slog.LevelError:
		level = "🔴 " + level
	default:
		level = "⚫️ " + level
	}
	fields := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	b := []byte{}
	if len(fields) > 0 {
		data, err := json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
		b = data
	}
	text := fmt.Sprintf("%s: %s\n\n%s", level, r.Message, b)
	_, err := h.bot.Write([]byte(text))
	return err
}
