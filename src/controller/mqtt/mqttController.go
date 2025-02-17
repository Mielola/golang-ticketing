package controllers

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttController struct {
	client   mqtt.Client
	broker   string
	clientID string
	topic    string
}

func NewMqttController(broker, clientID, topic string) *MqttController {
	controller := &MqttController{
		broker:   broker,
		clientID: clientID,
		topic:    topic,
	}
	controller.connect()
	return controller
}

func (mc *MqttController) connect() {
	// Set up MQTT client options
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mc.broker)
	opts.SetClientID(mc.clientID)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(1 * time.Minute)

	// Create the MQTT client
	mc.client = mqtt.NewClient(opts)

	// Connect to the broker
	if token := mc.client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %s\n", token.Error())
	}

	fmt.Println("Connected to MQTT broker")
}

func (mc *MqttController) SubscribeToTopic() {
	// Subscribe to a topic
	if token := mc.client.Subscribe(mc.topic, 0, mc.handleMessage); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic: %s\n", token.Error())
	}
	fmt.Println("Subscribed to topic:", mc.topic)
}

func (mc *MqttController) PublishMessage(message any) {
	// Publish message to the topic
	if token := mc.client.Publish(mc.topic, 0, true, message); token.Wait() && token.Error() != nil {
		log.Fatalf("Error publishing message: %s\n", token.Error())
	}
	fmt.Println("Message published:", message)
}

func (mc *MqttController) handleMessage(client mqtt.Client, msg mqtt.Message) {
	// Handle incoming messages
	fmt.Printf("Received message: %s on topic %s\n", msg.Payload(), msg.Topic())
}
