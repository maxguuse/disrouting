package disroute

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

const (
	TypeCommand               = discordgo.InteractionApplicationCommand
	TypeCommandAutocompletion = discordgo.InteractionApplicationCommandAutocomplete

	TypeSubcommand      = discordgo.ApplicationCommandOptionSubCommand
	TypeSubcommandGroup = discordgo.ApplicationCommandOptionSubCommandGroup

	TypeMessageComponent = discordgo.InteractionMessageComponent
)

type DiscordCmdOption = discordgo.ApplicationCommandInteractionDataOption

type OptionsMap map[string]*discordgo.ApplicationCommandInteractionDataOption

type HandlerFunc func(*Ctx) Response

type Router struct {
	session  *discordgo.Session
	handlers map[string]HandlerFunc
}

func New(token string) (*Router, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Router{
		session:  s,
		handlers: make(map[string]HandlerFunc),
	}, nil
}

func (r *Router) Handle(cmd *discordgo.ApplicationCommand, h HandlerFunc) {
	if _, ok := r.handlers[cmd.Name]; ok {
		return
	}

	r.handlers[cmd.Name] = h

	_, err := r.session.ApplicationCommandCreate(r.session.State.User.ID, "", cmd)
	if err != nil {
		panic(err)
	}
}

func (r *Router) InteractionHandler(_ *discordgo.Session, i *discordgo.InteractionCreate) {
	hd := buildHandlerData(i)

	ctx := newCtx(&ctxParams{
		Session: r.session,
		i:       i.Interaction,
		ctx:     context.Background(),
		Options: hd.opts,
	})

	h, ok := r.handlers[hd.path]
	if !ok {
		return
	}

	resp := h(ctx)

	if resp.Err != nil {
		r.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: resp.Err.Error(),
			},
		})

		return
	}

	if resp.CustomResponse != nil {
		r.session.InteractionRespond(i.Interaction, resp.CustomResponse)

		return
	}

	r.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: resp.Message,
		},
	})
}

func (r *Router) Session() *discordgo.Session {
	return r.session
}

func (r *Router) Mount(cmd *discordgo.ApplicationCommand) *SubRouter {
	return &SubRouter{
		root:        r,
		baseCmd:     cmd,
		basePath:    cmd.Name,
		lastOptions: &cmd.Options,
	}
}

type SubRouter struct {
	root        *Router
	baseCmd     *discordgo.ApplicationCommand
	basePath    string
	lastOptions *[]*discordgo.ApplicationCommandOption
}

func (r *SubRouter) Handle(cmd *discordgo.ApplicationCommandOption, h HandlerFunc) {
	if cmd.Type != discordgo.ApplicationCommandOptionSubCommand {
		return
	}

	path := r.basePath + ":" + cmd.Name
	r.root.handlers[path] = h

	*r.lastOptions = append(*r.lastOptions, cmd)

	_, err := r.root.session.ApplicationCommandCreate(r.root.session.State.User.ID, "", r.baseCmd)
	if err != nil {
		panic(err)
	}
}

func (r *SubRouter) Group(cmd *discordgo.ApplicationCommandOption) *SubRouter {
	if cmd.Type != discordgo.ApplicationCommandOptionSubCommandGroup {
		return r
	}

	*r.lastOptions = append(*r.lastOptions, cmd)

	return &SubRouter{
		root:        r.root,
		baseCmd:     r.baseCmd,
		basePath:    r.basePath + ":" + cmd.Name,
		lastOptions: &cmd.Options,
	}
}
