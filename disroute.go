package disroute

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

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

func (r *Router) Session() *discordgo.Session {
	return r.session
}

func (r *Router) Handle(cmd *discordgo.ApplicationCommand, h HandlerFunc) {
	if _, ok := r.handlers[cmd.Name]; ok {
		return
	}

	r.handlers[cmd.Name] = h

	_, err := r.session.ApplicationCommandCreate(r.session.State.Application.ID, "", cmd)
	if err != nil {
		panic(err)
	}
}

func (r *Router) InteractionHandler(_ *discordgo.Session, i *discordgo.InteractionCreate) {
	// TODO Get handler and path from somewhere

	ctx := newCtx(&ctxParams{
		Session: r.session,
		i:       i.Interaction,
		ctx:     context.Background(),
		Options: makeOptionsMap(i.ApplicationCommandData().Options), // TOOD Calculate options depth based on path
	})

	h, ok := r.handlers[i.ApplicationCommandData().Name]
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
	}

	r.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: resp.Message,
		},
	})
}

func makeOptionsMap(options []*discordgo.ApplicationCommandInteractionDataOption) OptionsMap {
	m := make(OptionsMap)

	for _, o := range options {
		m[o.Name] = o
	}

	return m
}
