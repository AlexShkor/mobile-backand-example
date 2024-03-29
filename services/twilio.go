package twilio

import (
	"fmt"

	"bitbucket.org/AlexShkor/cozytime/settings"
	twilio "github.com/carlosdp/twiliogo"
)

func SendCode(code string, phone string) error {
	conf := settings.Get()
	client := twilio.NewClient(conf.TwilioSID, conf.TwilioToken)
	message, err := twilio.NewMessage(client, "+18557007895", "+"+phone, twilio.Body("CozyTime! Your code: "+code))

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(message.Status)
		fmt.Println(message.Sid)
	}
	return err
}
