package iotf

import (
	"errors"
	"fmt"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/mqtt"
	"hub.jazz.net/git/bluemixgarage/lrsc-bridge/reporter"
	"math/rand"
	"regexp"
	"time"
)

type broker interface {
	connect() error
	statusReporter() reporter.StatusReporter
	publishMessageFromDevice(Event)
}

type iotfBroker struct {
	client mqtt.Client
	reporter.StatusReporter
	registrar deviceRegistrar
	commands  chan<- Command
}

const (
	deviceType = "LRSC"
)

func newClientOptions(credentials *Credentials, errChan chan<- error) mqtt.ClientOptions {
	return mqtt.ClientOptions{
		Broker:   fmt.Sprintf("tls://%v:%v", credentials.MqttHost, credentials.MqttSecurePort),
		ClientId: fmt.Sprintf("a:%v:$v", credentials.Org, generateClientIdSuffix()),
		Username: credentials.User,
		Password: credentials.Password,
		OnConnectionLost: func(err error) {
			logger.Error("IoTF connection lost handler called: " + err.Error())
			errChan <- errors.New("IoTF connection lost handler called: " + err.Error())
		},
	}
}

func generateClientIdSuffix() string {
	rand.Seed(time.Now().UTC().UnixNano())
	suffix := rand.Intn(1000)
	return string(suffix)
}

func newIoTFBroker(credentials *Credentials, commands chan<- Command, errChan chan<- error) *iotfBroker {
	clientOptions := newClientOptions(credentials, errChan)
	client := mqtt.NewPahoClient(clientOptions)
	registrar := iotfHttpRegistrar{credentials: credentials}
	reporter := reporter.New()
	return &iotfBroker{client: client, registrar: &registrar, commands: commands, StatusReporter: reporter}
}

func (self *iotfBroker) connect() error {
	var err error
	err = self.client.Start()
	if err != nil {
		self.Report("CONNECTION", err.Error())
		return err
	}

	self.Report("CONNECTION", "OK")

	logger.Info("Connected to MQTT")
	err = self.subscribeToCommandMessages(self.commands)
	if err != nil {
		self.Report("SUBSCRIPTION", err.Error())
		return err
	}
	self.Report("SUBSCRIPTION", "OK")
	return nil
}

func (self *iotfBroker) statusReporter() reporter.StatusReporter {
	return self.StatusReporter
}

func (self *iotfBroker) publishMessageFromDevice(event Event) {
	topic := fmt.Sprintf("iot-2/type/%v/id/%v/evt/TEST/fmt/json", deviceType, event.Device)
	logger.Debug("publishing event on topic %v: %v", topic, event)
	self.client.PublishMessage(topic, []byte(event.Payload))
}

func (self *iotfBroker) subscribeToCommandMessages(commands chan<- Command) error {
	topic := fmt.Sprintf("iot-2/type/%s/id/+/cmd/+/fmt/json", deviceType)
	return self.client.StartSubscription(topic, func(message mqtt.Message) {
		device := extractDeviceFromCommandTopic(message.Topic())
		command := Command{Device: device, Payload: string(message.Payload())}
		logger.Debug("received command message for %v", command.Device)
		commands <- command
	})
}

func extractDeviceFromCommandTopic(topic string) string {
	topicMatcher := regexp.MustCompile(`^iot-2/type/.*?/id/(.*?)/`)
	return topicMatcher.FindStringSubmatch(topic)[1]
}
