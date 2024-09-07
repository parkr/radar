package radar

import (
	"encoding/json"

	"mvdan.cc/xurls"
)

func ParseMailgunMessage(payload []byte) (RadarItem, error) {
	var m MailgunMessage
	if err := json.Unmarshal(payload, &m); err != nil {
		return RadarItem{}, err
	}

	// Parse the message body for URLs
	urls := xurls.Relaxed().FindAllString(m.StrippedText, -1)

	// If there are no URLs, we don't care about this message.
	if len(urls) == 0 {
		return RadarItem{}, nil
	}

	// Create the radar item
	// TODO: do we have a standard? Is this JSON? x-www-form-urlencoded?
	ri := RadarItem{}
	return ri, nil
}
