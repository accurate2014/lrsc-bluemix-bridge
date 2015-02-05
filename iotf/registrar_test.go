package iotf

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"net/http"
)

var _ = Describe("Registrar", func() {
	var (
		server           *ghttp.Server
		registrar        deviceRegistrar
		registrationPath string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		credentials := Credentials{
			User:     "testuser",
			Password: "testpass",
			Org:      "testorg",
			BaseUri:  server.URL()}

		registrationPath = "/organizations/testorg/devices"
		registrar = newIotfHttpRegistrar(&credentials, "test")

	})

	AfterEach(func() {
		server.Close()
	})

	Describe("registerDevice", func() {
		It("sends credentials", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", registrationPath),
					ghttp.VerifyBasicAuth("testuser", "testpass"),
					ghttp.RespondWith(http.StatusCreated, nil, nil),
				),
			)
			err := registrar.registerDevice("")
			Expect(err).To(Succeed())
		})

		It("POSTs the device information", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", registrationPath),
					ghttp.VerifyJSON(`{"id":"123456789", "type": "test"}`),
					ghttp.RespondWith(http.StatusCreated, nil, nil),
				),
			)
			err := registrar.registerDevice("123456789")
			Expect(err).To(Succeed())
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		Context("the device is not in IoTF", func() {
			It("succeeds", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.VerifyJSON(`{"id":"123456789", "type": "test"}`),
						ghttp.RespondWith(http.StatusCreated, nil, nil),
					),
				)
				err := registrar.registerDevice("123456789")
				Expect(err).To(Succeed())
			})
		})

		Context("the device already exists in IoTF", func() {
			It("succeeds", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.VerifyJSON(`{"id":"123456789", "type": "test"}`),
						ghttp.RespondWith(http.StatusConflict, nil, nil),
					),
				)
				err := registrar.registerDevice("123456789")
				Expect(err).To(Succeed())
			})
		})

		Context("the IoTF service is broken", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.RespondWith(http.StatusInternalServerError, nil, nil),
					),
				)
				err := registrar.registerDevice("")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("device has already been seen", func() {
			It("does not make API call", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.VerifyJSON(`{"id":"seen", "type": "test"}`),
						ghttp.RespondWith(http.StatusCreated, nil, nil),
					),
				)
				registrar.registerDevice("seen")
				registrar.registerDevice("seen")
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})

	})

	Describe("deviceRegistered", func() {
		Context("the device has previously been registered", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", registrationPath),
						ghttp.VerifyJSON(`{"id":"123456789", "type": "test"}`),
						ghttp.RespondWith(http.StatusCreated, nil, nil),
					),
				)
				registrar.registerDevice("123456789")
			})

			It("returns true", func() {
				Expect(registrar.deviceRegistered("123456789")).To(BeTrue())
			})
		})

		Context("the device has not previously been registered", func() {
			It("returns false", func() {
				Expect(registrar.deviceRegistered("123456789")).To(BeFalse())
			})
		})
	})
})
