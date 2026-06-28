package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"time"

	"github.com/bwmarrin/discordgo"
)

func banUser(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	config := settings.Get()
	for id, channel := range config.Channels {
		if m.ChannelID != id {
			continue
		}

		member, err := s.GuildMember(m.GuildID, m.Author.ID)
		if err != nil {
			slog.Error(
				"failed to fetch user roles",
				slog.String("message_id", m.ID),
				slog.String("user_id", m.Author.ID),
				slog.String("user_name", m.Author.DisplayName()),
				slog.Any("error", err),
			)
			return
		}

		for _, role := range member.Roles {
			if slices.Contains(channel.IgnoredRoles, role) {
				slog.Info(
					"user role is ignored",
					slog.String("message_id", m.ID),
					slog.String("user_id", m.Author.ID),
					slog.String("user_name", m.Author.DisplayName()),
					slog.String("role", role),
				)
				return
			}
		}

		days := int(channel.Delete.Duration / (24 * time.Hour))
		slog.Info(
			"banning user",
			slog.String("message_id", m.ID),
			slog.Bool("soft", channel.Soft),
			slog.String("user_id", m.Author.ID),
			slog.String("user_name", m.Author.DisplayName()),
			slog.String("delete", channel.Delete.Duration.String()),
			slog.Int("delete_days", days),
		)
		if err := s.GuildBanCreate(
			m.GuildID,
			m.Author.ID,
			days,
		); err != nil {
			slog.Error(
				"failed to ban user",
				slog.String("message_id", m.ID),
				slog.String("user_id", m.Author.ID),
				slog.String("user_name", m.Author.DisplayName()),
				slog.Any("error", err),
			)
		}

		if channel.Soft {
			if err := s.GuildBanDelete(
				m.GuildID,
				m.Author.ID,
			); err != nil {
				slog.Error(
					"failed to unban user",
					slog.String("message_id", m.ID),
					slog.String("user_id", m.Author.ID),
					slog.String("user_name", m.Author.DisplayName()),
					slog.Any("error", err),
				)
			}
		}
		return
	}
}

func runCtx(ctx context.Context) error {
	if err := settings.Load(); err != nil {
		slog.Error(
			"failed to read settings file",
			slog.Any("error", err),
		)
		return err
	}
	go settings.Watch(ctx)

	session, _ := discordgo.New("Bot " + settings.Get().Token)
	session.Identify.Intents |= discordgo.IntentGuildModeration
	session.AddHandler(banUser)

	if err := session.Open(); err != nil {
		slog.Error(
			"cannot open session",
			slog.Any("error", err),
		)
		return err
	}

	slog.Info(
		"opened session",
		slog.Any("user_id", session.State.User.ID),
		slog.Any("user_name", session.State.User.DisplayName()),
	)

	<-ctx.Done()
	slog.Info(
		"shutting down",
	)
	return session.Close()
}
func run() error {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for range 3 {
			<-c
			cancel(fmt.Errorf("interrupted by user"))
		}
		slog.Error("received 3 interrupts, hard exiting")
		os.Exit(1)
	}()

	return runCtx(ctx)
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}
