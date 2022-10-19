package radar

import (
	"testing"

	mailgun "github.com/mailgun/mailgun-go/v4"
)

func TestSendReply(t *testing.T) {
	mgMockServer := mailgun.NewMockServer()
	defer mgMockServer.Stop()
	mg := mailgun.NewMailgun("example.com", "faketestapikey")
	mg.SetAPIBase(mgMockServer.URL())
	svc := NewMailgunService(mg, "fromtest@example.com")

	err := svc.SendReply(createRequest{
		fromEmail: "fromtest@example.com",
		messageID: "123abc",
		subject:   "My cool radar item",
		url:       "http://example.com/?thebest=true",
	}, "Well done, Bob! Got it.")

	if err != nil {
		t.Fatal(err)
	}
}
