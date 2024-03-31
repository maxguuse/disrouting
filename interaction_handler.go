package disroute

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

func (r *Router) InteractionHandler(_ *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionPing || i.Type == discordgo.InteractionModalSubmit {
		return
	}

	ctx := newCtx(&ctxParams{
		Session: r.session,
		i:       i.Interaction,
		ctx:     context.Background(),
		Options: nil,
	})

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		r.handleCommand(ctx)
	case discordgo.InteractionApplicationCommandAutocomplete:
		r.handleAutocomplete(ctx)
	case discordgo.InteractionMessageComponent:
		r.handleComponent(ctx)
	}
}

func (r *Router) handleCommand(ctx *Ctx) {
	hd := buildHandlerData(ctx.Interaction())
	ctx.Options = hd.opts

	h, ok := r.handlers[hd.path]
	if !ok {
		return
	}

	resp := h(ctx)

	r.responseHandler(ctx, &resp)
}

func (r *Router) handleAutocomplete(ctx *Ctx) {
	hd := buildHandlerData(ctx.Interaction())
	ctx.Options = hd.opts

	h, ok := r.autocomp[hd.path]
	if !ok {
		return
	}

	resp := h(ctx)

	_ = r.ir.InteractionRespond(ctx.Interaction(), &discordgo.InteractionResponse{ // Add logger to router struct and log errors
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: resp,
		},
	})
}

func (r *Router) handleComponent(ctx *Ctx) {
	path := r.componentKeyFunc(ctx.Interaction())

	h, ok := r.components[path]
	if !ok {
		return
	}

	resp := h(ctx)

	r.responseHandler(ctx, &resp)
}
