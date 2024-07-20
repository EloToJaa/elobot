package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var CommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"single-autocomplete": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			data := i.ApplicationCommandData()
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf(
						"You picked %q autocompletion",
						data.Options[0].StringValue(),
					),
				},
			})
			if err != nil {
				panic(err)
			}
		// Autocomplete options introduce a new interaction type (8) for returning custom autocomplete results.
		case discordgo.InteractionApplicationCommandAutocomplete:
			data := i.ApplicationCommandData()
			choices := []*discordgo.ApplicationCommandOptionChoice{
				{
					Name:  "Autocomplete",
					Value: "autocomplete",
				},
				{
					Name:  "Autocomplete is best!",
					Value: "autocomplete_is_best",
				},
				{
					Name:  "Choice 3",
					Value: "choice3",
				},
				{
					Name:  "Choice 4",
					Value: "choice4",
				},
				{
					Name:  "Choice 5",
					Value: "choice5",
				},
				// And so on, up to 25 choices
			}

			if data.Options[0].StringValue() != "" {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  data.Options[0].StringValue(), // To get user input you just get value of the autocomplete option.
					Value: "choice_custom",
				})
			}

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionApplicationCommandAutocompleteResult,
				Data: &discordgo.InteractionResponseData{
					Choices: choices, // This is basically the whole purpose of autocomplete interaction - return custom options to the user.
				},
			})
			if err != nil {
				panic(err)
			}
		}
	},
}
