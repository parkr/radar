package radar

import (
	"context"
	"errors"

	mailgun "github.com/mailgun/mailgun-go/v3"
	"github.com/technoweenie/grohl"
)

var errNoFromEmail = errors.New("no from email was specified for mailgun")
var errMailgunNotSetup = errors.New("mailgun service isn't setup")

// NewMailgunService creates a new mailgun service which uses the given domain/credentials.
func NewMailgunService(mg mailgun.Mailgun, fromEmail string) MailgunService {
	return MailgunService{mg: mg, fromEmail: fromEmail}
}

// MailgunService provides a way to reply to incoming messages from Mailgun.
type MailgunService struct {
	mg mailgun.Mailgun

	fromEmail string
}

// SendReply sends a reply to the incoming request with the given body
func (svc MailgunService) SendReply(incoming createRequest, body string) error {
	if svc.fromEmail == "" {
		return errNoFromEmail
	}
	if svc.mg == nil {
		return errMailgunNotSetup
	}
	message := svc.mg.NewMessage(
		svc.fromEmail,
		"RE: "+incoming.subject,
		body,
		incoming.fromEmail)
	message.AddHeader("In-Reply-To", incoming.messageID)
	message.AddHeader("References", incoming.messageID)
	resp, id, err := svc.mg.Send(context.Background(), message)
	grohl.Log(grohl.Data{"id": id})
	Printf("ID: %s Resp: %s\n", id, resp)
	return err
}
