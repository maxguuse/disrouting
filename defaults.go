package disroute

import "github.com/bwmarrin/discordgo"

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

var defaultComponentKeyFunc = func(i *discordgo.Interaction) string {
	return i.MessageComponentData().CustomID
}
