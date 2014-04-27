package api_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/winston-ci/redgreen/api"
	"github.com/winston-ci/redgreen/api/builds"
)

var _ = Describe("API", func() {
	var proleServer *ghttp.Server

	var server *httptest.Server
	var client *http.Client

	BeforeEach(func() {
		proleServer = ghttp.NewServer()

		handler := api.New(log.New(GinkgoWriter, "test", 0), proleServer.URL())

		server = httptest.NewServer(handler)
		client = &http.Client{
			Transport: &http.Transport{},
		}
	})

	buildPayload := func(build *builds.Build) string {
		payload, err := json.Marshal(build)
		Ω(err).ShouldNot(HaveOccurred())

		return string(payload)
	}

	Describe("POST /builds", func() {
		var build *builds.Build
		var requestBody string
		var response *http.Response

		BeforeEach(func() {
			build = &builds.Build{
				Image:  "ubuntu",
				Script: "ls -al /",
			}

			requestBody = buildPayload(build)
		})

		JustBeforeEach(func() {
			var err error

			response, err = client.Post(
				server.URL+"/builds",
				"application/json",
				bytes.NewBufferString(requestBody),
			)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("returns 201", func() {
			Ω(response.StatusCode).Should(Equal(http.StatusCreated))
		})

		It("returns the build with an added guid and created_at", func() {
			var returnedBuild builds.Build

			err := json.NewDecoder(response.Body).Decode(&returnedBuild)
			Ω(err).ShouldNot(HaveOccurred())

			buildWithGuid := *build
			buildWithGuid.Guid = returnedBuild.Guid
			buildWithGuid.CreatedAt = returnedBuild.CreatedAt

			Ω(returnedBuild).Should(Equal(buildWithGuid))
			Ω(returnedBuild.CreatedAt.UnixNano()).Should(BeNumerically("~", time.Now().UnixNano(), time.Second))
		})

		Context("when image is omitted", func() {
			BeforeEach(func() {
				build.Image = ""
				requestBody = buildPayload(build)
			})

			It("returns 400", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusBadRequest))
			})
		})

		Context("when the payload is malformed JSON", func() {
			BeforeEach(func() {
				requestBody = "ß"
			})

			It("returns 400", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("GET /builds", func() {
		var response *http.Response
		var receivedBuilds []*builds.Build

		JustBeforeEach(func() {
			var err error

			response, err = client.Get(server.URL + "/builds")
			Ω(err).ShouldNot(HaveOccurred())

			err = json.NewDecoder(response.Body).Decode(&receivedBuilds)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("with no builds", func() {
			It("returns 200", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusOK))
			})

			It("returns an empty set of builds", func() {
				Ω(receivedBuilds).Should(BeEmpty())
			})
		})

		Context("with multiple builds", func() {
			var expectedBuilds [3]builds.Build

			BeforeEach(func() {
				expectedBuilds = [3]builds.Build{}

				for i := 0; i < 3; i++ {
					build := builds.Build{Image: "ubuntu"}

					response, err := client.Post(
						server.URL+"/builds",
						"application/json",
						bytes.NewBufferString(buildPayload(&build)),
					)
					Ω(err).ShouldNot(HaveOccurred())

					err = json.NewDecoder(response.Body).Decode(&build)
					Ω(err).ShouldNot(HaveOccurred())

					expectedBuilds[i] = build
				}
			})

			It("returns 200", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusOK))
			})

			It("returns them with the most recently created build first", func() {
				Ω(receivedBuilds).Should(HaveLen(3))
				Ω(receivedBuilds[0].Guid).Should(Equal(expectedBuilds[2].Guid))
				Ω(receivedBuilds[1].Guid).Should(Equal(expectedBuilds[1].Guid))
				Ω(receivedBuilds[2].Guid).Should(Equal(expectedBuilds[0].Guid))
			})
		})
	})
})
