package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/maxguuse/disroute"
)

func main() {
	router, err := disroute.New("token.must.be.here")
	if err != nil {
		log.Fatalln(err)
	}

	router.Session().AddHandler(router.InteractionHandler)

	router.Use(func(hf disroute.HandlerFunc) disroute.HandlerFunc {
		return func(c *disroute.Ctx) disroute.Response {
			log.Println(c.Interaction().Member.User.GlobalName)

			return hf(c)
		}
	})

	router.Handle(&discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Ping!",
	}, func(c *disroute.Ctx) disroute.Response {
		return disroute.Response{
			Message: "Pong!",
		}
	})

	AddSubPingRoute(router)

	if err = router.Open(); err != nil {
		log.Fatalln(err)
	}

	<-make(chan struct{})
}

func AddSubPingRoute(r *disroute.Router) {
	router := r.Mount(&discordgo.ApplicationCommand{
		Name:        "subping",
		Description: "Ping!",
	}).With(func(hf disroute.HandlerFunc) disroute.HandlerFunc {
		return func(c *disroute.Ctx) disroute.Response {
			log.Println("Subping router")

			return hf(c)
		}
	})

	router.Handle(&discordgo.ApplicationCommandOption{
		Name:        "ping",
		Description: "Ping!",
		Type:        discordgo.ApplicationCommandOptionSubCommand,
	}, func(c *disroute.Ctx) disroute.Response {
		return disroute.Response{
			Message: "Pong!",
		}
	})

	router.Handle(&discordgo.ApplicationCommandOption{
		Name:        "anotherping",
		Description: "Ping!",
		Type:        discordgo.ApplicationCommandOptionSubCommand,
	}, func(c *disroute.Ctx) disroute.Response {
		return disroute.Response{
			Message: "Pong!",
		}
	})

	grrouter := router.Group(&discordgo.ApplicationCommandOption{
		Name:        "gr",
		Description: "Ping!",
		Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
	}).With(func(hf disroute.HandlerFunc) disroute.HandlerFunc {
		return func(c *disroute.Ctx) disroute.Response {
			log.Println("Group router")

			return hf(c)
		}
	})

	grrouter.Handle(&discordgo.ApplicationCommandOption{
		Name:        "ping",
		Description: "Ping!",
		Type:        discordgo.ApplicationCommandOptionSubCommand,
	}, func(c *disroute.Ctx) disroute.Response {
		return disroute.Response{
			Message: "Pong!",
		}
	})
}
