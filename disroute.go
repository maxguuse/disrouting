package disroute

import (
	"github.com/bwmarrin/discordgo"
)

type Router struct {
	session *discordgo.Session

	cmds *[]*discordgo.ApplicationCommand

	middlewares []MiddlewareFunc

	handlers   map[string]HandlerFunc
	autocomp   map[string]AutocompleteFunc
	components map[string]HandlerFunc

	responseHandler     func(*Ctx, *Response)
	autocompleteHandler func(*Ctx, []*discordgo.ApplicationCommandOptionChoice)
	componentKeyFunc    func(*discordgo.Interaction) string
}

func New(token string) (*Router, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	cmds := make([]*discordgo.ApplicationCommand, 0, 100)

	return &Router{
		session: s,

		cmds: &cmds,

		handlers:   make(map[string]HandlerFunc),
		autocomp:   make(map[string]AutocompleteFunc),
		components: make(map[string]HandlerFunc),

		responseHandler:     defaultResponseHandler,
		autocompleteHandler: defaultAutocompleteHandler,
		componentKeyFunc:    defaultComponentKeyFunc,
	}, nil
}

func (r *Router) Open() error {
	err := r.session.Open()
	if err != nil {
		return err
	}

	_, err = r.session.ApplicationCommandBulkOverwrite(r.session.State.User.ID, "", *r.cmds)
	if err != nil {
		return err
	}

	return nil
}

func (r *Router) SetResponseHandler(h func(*Ctx, *Response)) {
	r.responseHandler = h
}

func (r *Router) SetAutocompleteHandler(h func(*Ctx, []*discordgo.ApplicationCommandOptionChoice)) {
	r.autocompleteHandler = h
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
	*r.cmds = append(*r.cmds, cmd)

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

	*r.cmds = append(*r.cmds, cmd)

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
		session:             r.session,
		middlewares:         newMiddlewares,
		cmds:                r.cmds,
		handlers:            r.handlers,
		autocomp:            r.autocomp,
		components:          r.components,
		responseHandler:     r.responseHandler,
		autocompleteHandler: r.autocompleteHandler,
		componentKeyFunc:    r.componentKeyFunc,
	}
}
