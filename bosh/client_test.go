package bosh_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/c0-ops/goblob/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BOSH Client", func() {

	Describe("Start", func() {

		It("starts a job instance", func() {
			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name/jobs/some-job/0":
					Expect(r.Method).To(Equal("PUT"))
					Expect(r.Header.Get("Content-Type")).To(Equal("text/yaml"))
					Expect(r.URL.RawQuery).To(Equal("state=start"))

					username, password, ok := r.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(username).To(Equal("some-username"))
					Expect(password).To(Equal("some-password"))

					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusFound)
				case "/tasks/1":
					Expect(r.Method).To(Equal("GET"))

					username, password, ok := r.BasicAuth()
					Expect(ok).To(BeTrue())
					Expect(username).To(Equal("some-username"))
					Expect(password).To(Equal("some-password"))

					if callCount == 3 {
						w.Write([]byte(`{"state": "done"}`))
					} else {
						w.Write([]byte(`{"state": "processing"}`))
					}
					callCount++
				default:
					Fail("unexpected route")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:                 server.URL,
				Username:            "some-username",
				Password:            "some-password",
				TaskPollingInterval: time.Nanosecond,
			})

			err := client.Start("some-deployment-name", "some-job", 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(callCount).To(Equal(4))
		})

		Context("failure cases", func() {
			It("errors when the deployment name contains invalid URL characters", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("something bad happened"))
				}))
				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				err := client.Start("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError(ContainSubstring("unexpected response 500 Internal Server Error:\nsomething bad happened")))
			})

			It("errors when the deployment name contains invalid URL characters", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "http://example.com%%%%%%%%%",
					Username: "some-username",
					Password: "some-password",
				})

				err := client.Start("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("errors when the bosh URL is malformed", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "banana://example.com",
					Username: "some-username",
					Password: "some-password",
				})

				err := client.Start("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("errors when the redirect location is bad", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Location", "%%%%%%%%%%%")
					w.WriteHeader(http.StatusFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				err := client.Start("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("returns an error on a bogus response body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("More Info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
					return nil, errors.New("a bad read happened")
				})

				err := client.Start("some-deployment-name", "some-job", 0)
				Expect(err).To(MatchError("a bad read happened"))
			})
		})
	})

	Describe("GetTaskOutput", func() {
		var server *httptest.Server

		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/tasks/1/output"))
				Expect(r.URL.RawQuery).To(Equal("type=event"))
				Expect(r.Method).To(Equal("GET"))

				w.Write([]byte(`
				{"time": 0, "error": {"code": 100, "message": "some-error" }, "stage": "some-stage", "tags": [ "some-tag" ], "total": 1, "task": "some-task-guid", "index": 1, "state": "some-state", "progress": 0}
{"time": 1, "error": {"code": 100, "message": "some-error" }, "stage": "some-stage", "tags": [ "some-tag" ], "total": 1, "task": "some-task-guid", "index": 1, "state": "some-new-state", "progress": 0}
				`))
			}))
		})

		It("returns task event output for a given task", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			taskOutputs, err := client.GetTaskOutput(1)
			Expect(err).NotTo(HaveOccurred())
			Expect(taskOutputs).To(ConsistOf(
				bosh.TaskOutput{
					Time: 0,
					Error: bosh.TaskError{
						Code:    100,
						Message: "some-error",
					},
					Stage:    "some-stage",
					Tags:     []string{"some-tag"},
					Total:    1,
					Task:     "some-task-guid",
					Index:    1,
					State:    "some-state",
					Progress: 0,
				},
				bosh.TaskOutput{
					Time: 1,
					Error: bosh.TaskError{
						Code:    100,
						Message: "some-error",
					},
					Stage:    "some-stage",
					Tags:     []string{"some-tag"},
					Total:    1,
					Task:     "some-task-guid",
					Index:    1,
					State:    "some-new-state",
					Progress: 0,
				},
			))
		})

		Context("failure cases", func() {
			It("error on a malformed URL", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "%%%%%%%%",
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.GetTaskOutput(1)
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("error on an empty URL", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      "",
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.GetTaskOutput(1)
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("errors on an unexpected status code with a body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadGateway)
					w.Write([]byte("More Info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.GetTaskOutput(1)
				Expect(err).To(MatchError("unexpected response 502 Bad Gateway:\nMore Info"))
			})

			It("should error on a bogus response body", func() {
				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
					return nil, errors.New("a bad read happened")
				})

				_, err := client.GetTaskOutput(1)
				Expect(err).To(MatchError("a bad read happened"))
			})

			It("error on malformed JSON", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(`%%%%%%%%`))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.GetTaskOutput(1)
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})
})
