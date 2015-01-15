package iotf

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Iotf", func() {
	Describe("extractCredentials", func() {
		It("extracts valid credentials", func() {
			vcapServices := `{"iotf-service":[{"name":"iotf","label":"iotf-service","tags":["internet_of_things","ibm_created"],"plan":"iotf-service-free","credentials":{"iotCredentialsIdentifier":"a2g6k39sl6r5","mqtt_host":"br2ybi.messaging.internetofthings.ibmcloud.com","mqtt_u_port":1883,"mqtt_s_port":8883,"base_uri":"https://internetofthings.ibmcloud.com:443/api/v0001","org":"br2ybi","apiKey":"a-br2ybi-y0tc7vicym","apiToken":"AJIpvsdJ!a__nqR(TK"}}]}`

			creds, err := extractCredentials(vcapServices)
			Expect(err).NotTo(HaveOccurred())
			Expect(creds.User).To(Equal("a-br2ybi-y0tc7vicym"))
		})

		It("errors with empty VCAP_SERVICES", func() {
			vcapServices := "{}"

			_, err := extractCredentials(vcapServices)
			Expect(err).To(HaveOccurred())
		})

		It("errors with empty string", func() {
			vcapServices := ""

			_, err := extractCredentials(vcapServices)
			Expect(err).To(HaveOccurred())
		})

	})

	Describe("IoTFManager", func() {
		var (
			iotfManager *IoTFManager
			mockBroker  *mockBroker
		)
		BeforeEach(func() {
			mockBroker = newMockBroker()
			iotfManager = &IoTFManager{broker: mockBroker}
		})

		Describe("Connect", func() {
			It("calls connect on the broker", func() {
				iotfManager.Connect()
				Expect(mockBroker.connected).To(BeTrue())
			})
		})
		Describe("Loop", func() {
			It("publishes events on the broker", func() {
				events := make(chan Event)
				iotfManager.events = events
				go iotfManager.Loop()

				event := Event{Device: "device", Payload: "message"}
				eventRead := false
				select {
				case events <- event:
					eventRead = true
				case <-time.After(time.Millisecond * 1):
					eventRead = false
				}

				Expect(eventRead).To(BeTrue())
				Expect(len(mockBroker.events)).To(Equal(1))
				Expect(mockBroker.events[0]).To(Equal(event))
				close(events)
			})

		})
		Describe("Error", func() {})
	})
})

type mockBroker struct {
	connected bool
	events    []Event
}

func newMockBroker() *mockBroker {
	events := make([]Event, 0)
	return &mockBroker{events: events}
}

func (self *mockBroker) connect() error {
	self.connected = true
	return nil
}

func (self *mockBroker) publishMessageFromDevice(event Event) {
	self.events = append(self.events, event)
}

type mockRegistrar struct {
	fail    bool
	devices map[string]struct{}
}

func newMockRegistrar() *mockRegistrar {
	return &mockRegistrar{devices: make(map[string]struct{})}
}

func (self *mockRegistrar) registerDevice(deviceType, deviceId string) error {
	if self.fail {
		return errors.New("")
	}

	self.devices[deviceId] = struct{}{}
	return nil
}
