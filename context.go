package disroute

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type Ctx struct {
	Options OptionsMap

	session *discordgo.Session
	i       *discordgo.Interaction
	ctx     context.Context
}

type ctxParams struct {
	Session *discordgo.Session
	i       *discordgo.Interaction
	ctx     context.Context
	Options OptionsMap
}

func newCtx(p *ctxParams) *Ctx {
	return &Ctx{
		session: p.Session,
		i:       p.i,
		ctx:     p.ctx,
		Options: p.Options,
	}
}

func (c *Ctx) Session() *discordgo.Session {
	return c.session
}

func (c *Ctx) Interaction() *discordgo.Interaction {
	return c.i
}

func (c *Ctx) Context() context.Context {
	return c.ctx
}

type Response struct {
	Message string
	Err     error
}
