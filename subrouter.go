package disroute

import "github.com/bwmarrin/discordgo"

type SubRouter struct {
	root        *Router
	baseCmd     *discordgo.ApplicationCommand
	basePath    string
	lastOptions *[]*discordgo.ApplicationCommandOption
	middlewares []MiddlewareFunc
}

func (r *SubRouter) Handle(cmd *discordgo.ApplicationCommandOption, h HandlerFunc) *AutocompletionBundle {
	if cmd.Type != discordgo.ApplicationCommandOptionSubCommand {
		return nil
	}

	path := r.basePath + ":" + cmd.Name

	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}

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

	newMiddlewares := make([]MiddlewareFunc, len(r.middlewares))
	copy(newMiddlewares, r.middlewares)

	return &SubRouter{
		root:        r.root,
		baseCmd:     r.baseCmd,
		basePath:    r.basePath + ":" + cmd.Name,
		lastOptions: &cmd.Options,
		middlewares: newMiddlewares,
	}
}

func (r *SubRouter) Use(mw1 MiddlewareFunc, mw ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, mw1)
	r.middlewares = append(r.middlewares, mw...)
}

func (r *SubRouter) With(mw1 MiddlewareFunc, mw ...MiddlewareFunc) *SubRouter {
	newMiddlewares := make([]MiddlewareFunc, len(r.middlewares), len(r.middlewares)+len(mw)+1)
	copy(newMiddlewares, r.middlewares)
	newMiddlewares = append(newMiddlewares, mw1)
	newMiddlewares = append(newMiddlewares, mw...)

	return &SubRouter{
		root:        r.root,
		baseCmd:     r.baseCmd,
		basePath:    r.basePath,
		lastOptions: r.lastOptions,
		middlewares: newMiddlewares,
	}
}
