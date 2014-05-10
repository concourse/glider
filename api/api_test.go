package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"

	ProleBuilds "github.com/winston-ci/prole/api/builds"
	"github.com/winston-ci/redgreen/api"
	"github.com/winston-ci/redgreen/api/builds"
)

var _ = Describe("API", func() {
	var proleServer *ghttp.Server

	var server *httptest.Server
	var client *http.Client

	BeforeEach(func() {
		proleServer = ghttp.NewServer()

		handler, err := api.New(log.New(GinkgoWriter, "test", 0), "peer-addr", proleServer.URL())
		Ω(err).ShouldNot(HaveOccurred())

		server = httptest.NewServer(handler)
		client = &http.Client{
			Transport: &http.Transport{},
		}
	})

	AfterEach(func() {
		server.Close()
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

		Ω(response.StatusCode).Should(Equal(http.StatusCreated))

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
					Env: [][2]string{
						{"FOO", "bar"},
					},
				})
			})

			BeforeEach(func() {
				proleServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/builds"),
						ghttp.VerifyJSONRepresenting(ProleBuilds.Build{
							Guid: build.Guid,

							Image: "ubuntu",
							Env: [][2]string{
								{"FOO", "bar"},
							},
							Script: "ls -al /",

							LogsURL:  "ws://peer-addr/builds/" + build.Guid + "/log/input",
							Callback: "http://peer-addr/builds/" + build.Guid + "/result",

							Source: ProleBuilds.BuildSource{
								Type: "raw",
								URI:  "http://peer-addr/builds/" + build.Guid + "/bits",
								Path: "some/path",
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
					},
				)
			})

			It("returns 200", func() {
				Ω(response.StatusCode).Should(Equal(http.StatusOK))
			})

			It("updates the build's status", func() {
				response, err := client.Get(endpoint)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response.StatusCode).Should(Equal(http.StatusOK))

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

	Describe("/builds/:guid/log/input", func() {
		var build builds.Build
		var endpoint string

		var conn *websocket.Conn
		var response *http.Response

		BeforeEach(func() {
			build = builds.Build{
				Guid: "some-guid",
			}

			endpoint = fmt.Sprintf(
				"ws://%s/builds/%s/log/input",
				server.Listener.Addr().String(),
				build.Guid,
			)
		})

		Context("with a valid build guid", func() {
			BeforeEach(func() {
				build = createBuild(
					builds.Build{
						Image:  "ubuntu",
						Script: "ls -al /",
					},
				)

				endpoint = fmt.Sprintf(
					"ws://%s/builds/%s/log/input",
					server.Listener.Addr().String(),
					build.Guid,
				)
			})

			It("returns 101", func() {
				_, response, err := websocket.DefaultDialer.Dial(endpoint, nil)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(response.StatusCode).Should(Equal(http.StatusSwitchingProtocols))
			})

			Context("when messages are written", func() {
				BeforeEach(func() {
					var err error

					conn, response, err = websocket.DefaultDialer.Dial(endpoint, nil)
					Ω(err).ShouldNot(HaveOccurred())

					err = conn.WriteMessage(websocket.BinaryMessage, []byte("hello1"))
					Ω(err).ShouldNot(HaveOccurred())

					err = conn.WriteMessage(websocket.BinaryMessage, []byte("hello2\n"))
					Ω(err).ShouldNot(HaveOccurred())

					err = conn.WriteMessage(websocket.BinaryMessage, []byte("hello3"))
					Ω(err).ShouldNot(HaveOccurred())
				})

				outputSink := func() *gbytes.Buffer {
					outEndpoint := fmt.Sprintf(
						"ws://%s/builds/%s/log/output",
						server.Listener.Addr().String(),
						build.Guid,
					)

					outConn, outResponse, err := websocket.DefaultDialer.Dial(outEndpoint, nil)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(outResponse.StatusCode).Should(Equal(http.StatusSwitchingProtocols))

					buf := gbytes.NewBuffer()

					go func() {
						for {
							_, msg, err := outConn.ReadMessage()
							if err != nil {
								break
							}

							buf.Write(msg)
						}
					}()

					return buf
				}

				It("presents them to /builds/{guid}/logs/output", func() {
					Eventually(outputSink()).Should(gbytes.Say("hello1hello2\nhello3"))
				})

				It("streams them to all open connections to /build/{guid}/logs/output", func() {
					sink1 := outputSink()
					sink2 := outputSink()

					err := conn.WriteMessage(websocket.BinaryMessage, []byte("some message"))
					Ω(err).ShouldNot(HaveOccurred())

					Eventually(sink1).Should(gbytes.Say("some message"))
					Eventually(sink2).Should(gbytes.Say("some message"))
				})
			})
		})

		Context("with an invalid build guid", func() {
			It("returns 404", func() {
				_, response, err := websocket.DefaultDialer.Dial(endpoint, nil)
				Ω(err).Should(HaveOccurred())

				Ω(response.StatusCode).Should(Equal(http.StatusNotFound))
			})
		})
	})
})
