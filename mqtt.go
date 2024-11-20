package radar

import (
	"context"
	"log"

	"gosrc.io/mqtt"
)

type EmailFromMqttHandler struct {
	Client        *mqtt.Client
	ClientManager *mqtt.ClientManager

	MQTTTopic string

	AllowedSenders []string

	Debug bool

	RadarItems RadarItemsStorageService

	CreateQueue chan createRequest

	RadarCreatedChan chan bool
}

// NewEmailFromMqttHandler creates a new handler that subscribes to mqtt and processes incoming emails from there.
// It's still expecting mailgun payloads, but it receives them from mqtt instead of http.
func NewEmailFromMqttHandler(radarItemsService RadarItemsStorageService, mqttConnStr string, mqttTopic string, allowedSenders []string, debug bool, radarCreatedChan chan bool) EmailFromMqttHandler {
	return EmailFromMqttHandler{
		Client:           mqtt.NewClient(mqttConnStr),
		MQTTTopic:        mqttTopic,
		AllowedSenders:   allowedSenders,
		Debug:            debug,
		RadarItems:       radarItemsService,
		CreateQueue:      make(chan createRequest, 10),
		RadarCreatedChan: radarCreatedChan,
	}
}

func (h *EmailFromMqttHandler) Start() {
	// Start the mqtt client
	h.Client.ClientID = "radar-mqtt-handler"
	messages := make(chan mqtt.Message)
	h.Client.Messages = messages

	postConnect := func(c *mqtt.Client) {
		log.Println("[mqtt] connected")
		topic := mqtt.Topic{Name: h.MQTTTopic, QOS: 0}
		c.Subscribe(topic)
	}

	h.ClientManager = mqtt.NewClientManager(h.Client, postConnect)
	h.ClientManager.Start()

	for m := range messages {
		h.OnMessage(m)
	}
}

func (h *EmailFromMqttHandler) OnMessage(m mqtt.Message) {
	log.Printf("[mqtt] received message on topic %s: %+v\n", m.Topic, m.Payload)
	radarItem, err := ParseMailgunMessage(m.Payload)
	if err != nil {
		log.Printf("[mqtt] error parsing message: %v\n", err)
		return
	}
	h.RadarItems.Create(context.Background(), radarItem)
}

func (h *EmailFromMqttHandler) Shutdown(ctx context.Context) error {
	h.ClientManager.Stop() // this also disconnects the client
	return nil
}
