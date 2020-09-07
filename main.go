/*
 *    Copyright © 2020 Haruka Network Development
 *    This file is part of Haruka X.
 *
 *    Haruka X is free software: you can redistribute it and/or modify
 *    it under the terms of the Raphielscape Public License as published by
 *    the Devscapes Open Source Holding GmbH., version 1.d
 *
 *    Haruka X is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    Devscapes Raphielscape Public License for more details.
 *
 *    You should have received a copy of the Devscapes Raphielscape Public License
 */

package main

import (
	"log"
	"os"
	"strconv"

	"github.com/NoodleSoup/NoodleX/noodlex/modules/rules"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/stickers"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/ud"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/NoodleSoup/NoodleX/noodlex"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/admin"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/bans"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/blacklist"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/deleting"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/feds"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/help"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/misc"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/muting"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/notes"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/sql"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/users"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/utils/caching"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/utils/error_handling"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/warns"
	"github.com/NoodleSoup/NoodleX/noodlex/modules/welcome"
	"github.com/PaulSonOfLars/gotgbot"
	"github.com/PaulSonOfLars/gotgbot/ext"
	"github.com/PaulSonOfLars/gotgbot/handlers"
)

func main() {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder

	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(cfg), os.Stdout, zap.InfoLevel))
	defer logger.Sync() // flushes buffer, if any
	// Create updater instance
	u, err := gotgbot.NewUpdater(logger, noodlex.BotConfig.ApiKey)
	error_handling.FatalError(err)

	// Add start handler
	u.Dispatcher.AddHandler(handlers.NewArgsCommand("start", start))

	// Create database tables if not already existing
	sql.EnsureBotInDb(u)

	// Prepare Caching Service
	caching.InitCache()
	//caching.InitRedis()

	// Add module handlers
	bans.LoadBans(u)
	users.LoadUsers(u)
	admin.LoadAdmin(u)
	warns.LoadWarns(u)
	misc.LoadMisc(u)
	muting.LoadMuting(u)
	deleting.LoadDelete(u)
	blacklist.LoadBlacklist(u)
	feds.LoadFeds(u)
	notes.LoadNotes(u)
	help.LoadHelp(u)
	welcome.LoadWelcome(u)
	rules.LoadRules(u)
	ud.LoadUd(u)
	stickers.LoadStickers(u)

	if noodlex.BotConfig.DropUpdate == "True" {
		log.Println("[Info][Core] Using Clean Long Polling")
		err = u.StartCleanPolling()
		error_handling.HandleErr(err)
	} else {
		log.Println("[Info][Core] Using Long Polling")
		err = u.StartPolling()
		error_handling.HandleErr(err)
	}

	u.Idle()
}

func start(_ ext.Bot, u *gotgbot.Update, args []string) error {
	msg := u.EffectiveMessage

	if u.EffectiveChat.Type == "private" {
		if len(args) != 0 {
			if _, err := strconv.Atoi(args[0][2:]); err == nil {
				chatRules := sql.GetChatRules(args[0])
				if chatRules != nil {
					_, err := msg.ReplyHTML(chatRules.Rules)
					return err
				}
				_, err := msg.ReplyText("The group admins haven't set any rules for this chat yet. This probably doesn't " +
					"mean it's lawless though!")
				log.Println(args[0])
				return err
			}
		}
	}

	_, err := msg.ReplyTextf("Hi there! I'm a telegram group management bot, written in Go." +
		"\nFor any questions or bug reports, you can head over to @NatalieSupport.")
	return err
}
