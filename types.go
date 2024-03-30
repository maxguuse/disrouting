package disroute

import "github.com/bwmarrin/discordgo"

type DiscordCmdOption = discordgo.ApplicationCommandInteractionDataOption

type OptionsMap map[string]*discordgo.ApplicationCommandInteractionDataOption

type HandlerFunc func(*Ctx) Response

type MiddlewareFunc func(HandlerFunc) HandlerFunc
