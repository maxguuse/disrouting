package disroute

import "github.com/bwmarrin/discordgo"

type AutocompleteFunc func(*Ctx) []*discordgo.ApplicationCommandOptionChoice

type AutocompletionBundle struct {
	router *Router
	path   string
}

func (b *AutocompletionBundle) WithAutocompletion(h AutocompleteFunc) {
	b.router.autocomp[b.path] = h
}
