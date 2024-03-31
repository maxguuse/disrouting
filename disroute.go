package disroute

import (
	"github.com/bwmarrin/discordgo"
)

type commandRegisterer interface {
	ApplicationCommandCreate(string, string, *discordgo.ApplicationCommand, ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error)
}

type commandRegistererImpl struct {
	*discordgo.Session
}

func (c *commandRegistererImpl) ApplicationCommandCreate(appID string, guildID string, cmd *discordgo.ApplicationCommand, options ...discordgo.RequestOption) (*discordgo.ApplicationCommand, error) {
	return c.Session.ApplicationCommandCreate(appID, guildID, cmd, options...)
}

type interactionResponder interface {
	InteractionRespond(*discordgo.Interaction, *discordgo.InteractionResponse, ...discordgo.RequestOption) error
}

type interactionResponderImpl struct {
	*discordgo.Session
}

func (i *interactionResponderImpl) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	return i.Session.InteractionRespond(interaction, resp, options...)
}

type Router struct {
	session *discordgo.Session

	cr commandRegisterer
	ir interactionResponder

	middlewares []MiddlewareFunc

	handlers   map[string]HandlerFunc
	autocomp   map[string]AutocompleteFunc
	components map[string]HandlerFunc

	componentKeyFunc func(*discordgo.Interaction) string
	responseHandler  func(*Ctx, *Response)
}

func New(token string) (*Router, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Router{
		session: s,

		cr: &commandRegistererImpl{s},
		ir: &interactionResponderImpl{s},

		handlers:   make(map[string]HandlerFunc),
		autocomp:   make(map[string]AutocompleteFunc),
		components: make(map[string]HandlerFunc),

		responseHandler:  defaultResponseHandler,
		componentKeyFunc: defaultComponentKeyFunc,
	}, nil
}

func (r *Router) SetResponseHandler(h func(*Ctx, *Response)) {
	r.responseHandler = h
}

func (r *Router) SetComponentKeyFunc(f func(*discordgo.Interaction) (key string)) {
	r.componentKeyFunc = f
}

func (r *Router) Session() *discordgo.Session {
	return r.session
}

func (r *Router) Handle(cmd *discordgo.ApplicationCommand, h HandlerFunc) *AutocompletionBundle {
	if _, ok := r.handlers[cmd.Name]; ok {
		return nil
	}

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}

	r.handlers[cmd.Name] = h

	_, err := r.cr.ApplicationCommandCreate(r.session.State.User.ID, "", cmd)
	if err != nil {
		panic(err)
	}

	return &AutocompletionBundle{
		router: r,
		path:   cmd.Name,
	}
}

func (r *Router) HandleComponent(key string, h HandlerFunc) {
	if _, ok := r.components[key]; ok {
		return
	}

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}

	r.components[key] = h
}

func (r *Router) Mount(cmd *discordgo.ApplicationCommand) *SubRouter {
	newMiddlewares := make([]MiddlewareFunc, len(r.middlewares))
	copy(newMiddlewares, r.middlewares)

	return &SubRouter{
		root:        r,
		baseCmd:     cmd,
		basePath:    cmd.Name,
		lastOptions: &cmd.Options,
		middlewares: newMiddlewares,
	}
}

func (r *Router) Use(mw1 MiddlewareFunc, mw ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mw1)
	r.middlewares = append(r.middlewares, mw...)
}

func (r *Router) With(mw1 MiddlewareFunc, mw ...MiddlewareFunc) *Router {
	newMiddlewares := make([]MiddlewareFunc, len(r.middlewares), len(r.middlewares)+len(mw)+1)
	copy(newMiddlewares, r.middlewares)
	newMiddlewares = append(newMiddlewares, mw1)
	newMiddlewares = append(newMiddlewares, mw...)

	return &Router{
		session:         r.session,
		middlewares:     newMiddlewares,
		handlers:        r.handlers,
		autocomp:        r.autocomp,
		responseHandler: r.responseHandler,
	}
}
