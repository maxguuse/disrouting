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
	autocomp map[string]AutocompleteFunc

	responseHandler func(*Ctx, *Response)
}

var defaultResponseHandler = func(ctx *Ctx, resp *Response) {
	if resp.Err != nil {
		_ = ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: resp.Err.Error(),
			},
		})

		return
	}

	if resp.CustomResponse != nil {
		_ = ctx.Session().InteractionRespond(ctx.Interaction(), resp.CustomResponse)

		return
	}

	_ = ctx.Session().InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: resp.Message,
		},
	})
}

func New(token string) (*Router, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Router{
		session:  s,
		handlers: make(map[string]HandlerFunc),
		autocomp: make(map[string]AutocompleteFunc),

		responseHandler: defaultResponseHandler,
	}, nil
}

func (r *Router) SetResponseHandler(h func(*Ctx, *Response)) {
	r.responseHandler = h
}

func (r *Router) Handle(cmd *discordgo.ApplicationCommand, h HandlerFunc) *AutocompletionBundle {
	if _, ok := r.handlers[cmd.Name]; ok {
		return nil
	}

	r.handlers[cmd.Name] = h

	_, err := r.session.ApplicationCommandCreate(r.session.State.User.ID, "", cmd)
	if err != nil {
		panic(err)
	}

	return &AutocompletionBundle{
		router: r,
		path:   cmd.Name,
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

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		r.handleCommand(hd, ctx)
	case discordgo.InteractionApplicationCommandAutocomplete:
		r.handleAutocomplete(hd, ctx)
	}

}

func (r *Router) handleCommand(hd *handlerData, ctx *Ctx) {
	h, ok := r.handlers[hd.path]
	if !ok {
		return
	}

	resp := h(ctx)

	r.responseHandler(ctx, &resp)
}

func (r *Router) handleAutocomplete(hd *handlerData, ctx *Ctx) {
	h, ok := r.autocomp[hd.path]
	if !ok {
		return
	}

	resp := h(ctx)

	_ = r.session.InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{ // Add logger to router struct and log errors
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: resp,
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

func (r *SubRouter) Handle(cmd *discordgo.ApplicationCommandOption, h HandlerFunc) *AutocompletionBundle {
	if cmd.Type != discordgo.ApplicationCommandOptionSubCommand {
		return nil
	}

	path := r.basePath + ":" + cmd.Name
	r.root.handlers[path] = h

	*r.lastOptions = append(*r.lastOptions, cmd)

	_, err := r.root.session.ApplicationCommandCreate(r.root.session.State.User.ID, "", r.baseCmd)
	if err != nil {
		panic(err)
	}

	return &AutocompletionBundle{
		router: r.root,
		path:   path,
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

type AutocompleteFunc func(*Ctx) []*discordgo.ApplicationCommandOptionChoice

type AutocompletionBundle struct {
	router *Router
	path   string
}

func (b *AutocompletionBundle) WithAutocompletion(h AutocompleteFunc) {
	b.router.autocomp[b.path] = h
}
