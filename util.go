package disroute

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type handlerData struct {
	path string
	opts map[string]*DiscordCmdOption
}

func buildHandlerData(i *discordgo.InteractionCreate) *handlerData {
	d := i.ApplicationCommandData()

	pathParts := []string{d.Name}
	options := buildOptionsMap(d.Options)

	if len(d.Options) == 0 {
		return &handlerData{
			path: strings.Join(pathParts, ":"),
			opts: options,
		}
	}

	if d.Options[0].Type == TypeSubcommand {
		pathParts = append(pathParts, d.Options[0].Name)
		options = buildOptionsMap(d.Options[0].Options)
	}

	if d.Options[0].Type == TypeSubcommandGroup {
		pathParts = append(pathParts,
			d.Options[0].Name,
			d.Options[0].Options[0].Name,
		)
		options = buildOptionsMap(d.Options[0].Options[0].Options)
	}

	return &handlerData{
		path: strings.Join(pathParts, ":"),
		opts: options,
	}
}

func buildOptionsMap(options []*DiscordCmdOption) map[string]*DiscordCmdOption {
	commandOptions := make(map[string]*DiscordCmdOption)
	for _, option := range options {
		commandOptions[option.Name] = option
	}

	return commandOptions
}
