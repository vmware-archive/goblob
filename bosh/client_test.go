package bosh_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	Describe("GetVMs", func() {
		It("retrieves the list of deployment VMs given a deployment name", func() {
			var taskCallCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username, password, ok := r.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				switch r.URL.Path {
				case "/deployments/some-deployment-name/vms":
					Expect(r.URL.RawQuery).To(Equal("format=full"))
					host, _, err := net.SplitHostPort(r.Host)
					Expect(err).NotTo(HaveOccurred())

					location := &url.URL{
						Scheme: "http",
						Host:   host,
						Path:   "/tasks/1",
					}

					w.Header().Set("Location", location.String())
					w.WriteHeader(http.StatusFound)
				case "/tasks/1":
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte(`{"state":"done"}`))
					taskCallCount++
				case "/tasks/1/output":
					Expect(r.URL.RawQuery).To(Equal("type=result"))
					Expect(taskCallCount).NotTo(Equal(0))

					w.Write([]byte(`
						{"index": 0, "job_name": "consul_z1", "job_state":"some-state", "ips": ["1.2.3.4"]}
						{"index": 0, "job_name": "etcd_z1", "job_state":"some-state", "ips": ["1.2.3.5"]}
						{"index": 1, "job_name": "etcd_z1", "job_state":"some-other-state", "ips": ["1.2.3.6"]}
						{"index": 2, "job_name": "etcd_z1", "job_state":"some-more-state", "ips": ["1.2.3.7"]}
					`))
				default:
					Fail("unknown route")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			vms, err := client.GetVMs("some-deployment-name")
			Expect(err).NotTo(HaveOccurred())
			Expect(vms).To(ConsistOf([]bosh.VM{
				{
					Index:   0,
					JobName: "consul_z1",
					State:   "some-state",
					IPs:     []string{"1.2.3.4"},
				},
				{
					Index:   0,
					JobName: "etcd_z1",
					State:   "some-state",
					IPs:     []string{"1.2.3.5"},
				},
				{
					Index:   1,
					JobName: "etcd_z1",
					State:   "some-other-state",
					IPs:     []string{"1.2.3.6"},
				},
				{
					Index:   2,
					JobName: "etcd_z1",
					State:   "some-more-state",
					IPs:     []string{"1.2.3.7"},
				},
			}))
		})

		Context("failure cases", func() {
			It("errors when the URL is malformed", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "http://%%%%%",
				})

				_, err := client.GetVMs("some-deployment-name")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("errors when the protocol scheme is invalid", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "banana://example.com",
				})

				_, err := client.GetVMs("some-deployment-name")
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
			})

			It("errors when checking the task fails", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/deployments/some-deployment-name/vms":
						w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
						w.WriteHeader(http.StatusFound)
					case "/tasks/1":
						w.Write([]byte("%%%"))
					default:
						Fail("unexpected route")
					}
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.GetVMs("some-deployment-name")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})

			It("should error on a non StatusFound status code with a body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name/vms"))
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("More Info"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.GetVMs("some-deployment-name")
				Expect(err).To(MatchError("unexpected response 404 Not Found:\nMore Info"))
			})

			It("errors when the redirect URL is malformed", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name/vms"))
					w.Header().Set("Location", "http://%%%%%/tasks/1")
					w.WriteHeader(http.StatusFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.GetVMs("some-deployment-name")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})

			It("should error on malformed JSON", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusFound)
					w.Write([]byte("%%%%%%\n%%%%%%%%%%%\n"))
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.GetVMs("some-deployment-name")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})

			It("should error on a bogus response body", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/deployments/some-deployment-name/vms":
						w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
						w.WriteHeader(http.StatusFound)
					case "/tasks/1":
						w.Write([]byte(`{"state": "done"}`))
					case "/tasks/1/output":
						w.Write([]byte(""))
					default:
						Fail("unexpected route")
					}
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
					return nil, errors.New("a bad read happened")
				})
				_, err := client.GetVMs("some-deployment-name")
				Expect(err).To(MatchError("a bad read happened"))
			})

			It("should error on a bogus response body when unexpected response occurs", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name/vms"))
					w.WriteHeader(http.StatusNotFound)
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
				_, err := client.GetVMs("some-deployment-name")
				Expect(err).To(MatchError("a bad read happened"))
			})
		})
	})
})
