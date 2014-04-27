package api_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/runtime-schema/models"
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

		handler := api.New(log.New(GinkgoWriter, "test", 0), "peer-addr", proleServer.URL())

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

	createBuild := func(build builds.Build) builds.Build {
		response, err := client.Post(
			server.URL+"/builds",
			"application/json",
			bytes.NewBufferString(buildPayload(&build)),
		)
		Ω(err).ShouldNot(HaveOccurred())

		err = json.NewDecoder(response.Body).Decode(&build)
		Ω(err).ShouldNot(HaveOccurred())

		return build
	}

	Describe("POST /builds", func() {
		var build *builds.Build
		var requestBody string
		var response *http.Response

		BeforeEach(func() {
			build = &builds.Build{
				Image:  "ubuntu",
				Path:   "some/path",
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
			var expectedBuilds []builds.Build

			BeforeEach(func() {
				expectedBuilds = []builds.Build{
					createBuild(builds.Build{Image: "image1"}),
					createBuild(builds.Build{Image: "image2"}),
					createBuild(builds.Build{Image: "image3"}),
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

	Describe("POST /builds/:guid/bits", func() {
		var build builds.Build

		var response *http.Response

		BeforeEach(func() {
			build = builds.Build{
				Guid: "some-guid",
			}
		})

		JustBeforeEach(func() {
			var err error

			// set up a consumer
			go client.Get(server.URL + "/builds/" + build.Guid + "/bits")

			response, err = client.Post(
				server.URL+"/builds/"+build.Guid+"/bits",
				"application/octet-stream",
				bytes.NewBufferString("streamed body"),
			)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("with a valid build guid", func() {
			BeforeEach(func() {
				build = createBuild(builds.Build{
					Image:  "ubuntu",
					Path:   "some/path",
					Script: "ls -al /",
					Environment: map[string]string{
						"FOO": "bar",
					},
				})
			})

			BeforeEach(func() {
				proleServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/builds"),
						ghttp.VerifyJSONRepresenting(builds.ProleBuild{
							Guid: build.Guid,

							LogConfig: models.LogConfig{
								Guid:       build.Guid,
								SourceName: "BLD",
							},

							Image:  "ubuntu",
							Script: "ls -al /",

							Source: builds.ProleBuildSource{
								Type: "raw",
								URI:  "http://peer-addr/builds/" + build.Guid + "/bits",
								Path: "some/path",
							},

							Callback: "http://peer-addr/builds/" + build.Guid + "/result",

							Parameters: map[string]string{
								"FOO": "bar",
							},
						}),
						ghttp.RespondWith(201, ""),
					),
				)
			})

			It("triggers a build and returns 201", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusCreated))
			})

			Context("when prole fails", func() {
				BeforeEach(func() {
					proleServer.SetHandler(0, ghttp.RespondWith(500, ""))
				})

				It("returns 500", func() {
					Ω(response.StatusCode).Should(Equal(http.StatusServiceUnavailable))
				})
			})
		})

		Context("with an invalid build guid", func() {
			It("returns 404", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("GET /builds/:guid/bits", func() {
		var build builds.Build

		var response *http.Response

		BeforeEach(func() {
			build = builds.Build{
				Guid: "some-guid",
			}
		})

		streamBits := func() {
			var err error

			response, err = client.Get(server.URL + "/builds/" + build.Guid + "/bits")
			Ω(err).ShouldNot(HaveOccurred())
		}

		JustBeforeEach(streamBits)

		Context("with a valid build guid", func() {
			BeforeEach(func() {
				build = createBuild(builds.Build{Image: "ubuntu"})
			})

			Context("with bits", func() {
				BeforeEach(func() {
					gotBits := &sync.WaitGroup{}
					gotBits.Add(1)

					proleServer.AppendHandlers(
						func(w http.ResponseWriter, req *http.Request) {
							gotBits.Done()
							w.WriteHeader(201)
						},
					)

					go func() {
						defer GinkgoRecover()

						_, err := client.Post(
							server.URL+"/builds/"+build.Guid+"/bits",
							"application/octet-stream",
							bytes.NewBufferString("streamed body"),
						)
						Ω(err).ShouldNot(HaveOccurred())
					}()

					gotBits.Wait()
				})

				It("returns 200", func() {
					Ω(response.StatusCode).Should(Equal(http.StatusOK))
				})

				It("streams the bits that are being uploaded", func() {
					body, err := ioutil.ReadAll(response.Body)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(string(body)).Should(Equal("streamed body"))
				})
			})

			Context("with no bits within 1 second", func() {
				var startedAt time.Time

				BeforeEach(func() {
					startedAt = time.Now()
				})

				It("returns 404", func() {
					Ω(response.StatusCode).Should(Equal(http.StatusNotFound))
					Ω(time.Since(startedAt)).Should(BeNumerically(">=", time.Second))
				})
			})
		})

		Context("with an invalid build guid", func() {
			It("returns 404", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})
	})

	Describe("GET/PUT /builds/:guid/result", func() {
		var build builds.Build
		var endpoint string

		var response *http.Response

		BeforeEach(func() {
			build = builds.Build{
				Guid: "some-guid",
			}
		})

		JustBeforeEach(func() {
			var err error

			endpoint = server.URL + "/builds/" + build.Guid + "/result"

			req, err := http.NewRequest("PUT", endpoint, nil)
			Ω(err).ShouldNot(HaveOccurred())

			reqPayload := bytes.NewBufferString(`{"status":"succeeded"}`)
			req.Header.Set("Content-Type", "application/json")
			req.Body = ioutil.NopCloser(reqPayload)

			response, err = client.Do(req)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("with a valid build guid", func() {
			BeforeEach(func() {
				build = createBuild(
					builds.Build{
						Image:  "ubuntu",
						Script: "ls -al /",
						Environment: map[string]string{
							"FOO": "bar",
						},
					},
				)
			})

			It("returns 200", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusOK))
			})

			It("updates the build's status", func() {
				response, err := client.Get(endpoint)
				Ω(err).ShouldNot(HaveOccurred())

				var result builds.BuildResult
				err = json.NewDecoder(response.Body).Decode(&result)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(result.Status).Should(Equal("succeeded"))
			})
		})

		Context("with an invalid build guid", func() {
			It("returns 404", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})
	})
})
