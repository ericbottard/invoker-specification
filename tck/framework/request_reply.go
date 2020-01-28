package framework

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"io/ioutil"
	"math"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"
)

var Request_reply = Suite{
	Name:        "rr",
	Description: "Request / Reply Interaction",
	Port:        8080,
	Cases: []*Testcase{
		/*
		{
			Name:           "rr-0000",
			Description:    "MUST start an http/2 server listening on $PORT",
			Image:          "upper",
			SetUpContainer: setUpContainerUsingPortEnvVar,
			TearDownContainer: func(container *Container, runner *Runner) {

			},
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				}
			},
		},
		*/
		{
			Name:        "rr-0001",
			Description: "MUST NOT reply on paths other than / or methods other than POST",
			Image:       "upper",
			T: func(port int) {
				response, err := http.Post(fmt.Sprintf("http://localhost:%d/bogus", port), "text/plain", strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if "HELLO" == string(result) {
					panic("The function function should only be exposed on /")
				}

				req, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				response, err = http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if "HELLO" == string(result) {
					panic("The function function should only be exposed on /")
				}
			},
		},
		{
			Name:        "rr-0002",
			Description: "MUST honor the Accept header",
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("Accept", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				} else if "HELLO" != string(result) {
					panic("Expected result as text/plain HELLO, got " + string(result))
				} else {
					hs := response.Header[http.CanonicalHeaderKey("Content-Type")]
					if len(hs) != 1 {
						panic("No Content-Type set on response")
					}
					mediaType, _, err := mime.ParseMediaType(hs[0])
					if err != nil {
						panic(fmt.Sprintf("Error parsing content-type: %v", err))
					} else if mediaType != "text/plain" {
						panic(fmt.Sprintf("Expected response Content-Type to be set to text/plain, got %v", mediaType))
					}
				}

				req, err = http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("Accept", "application/json")
				response, err = http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if result, err := ioutil.ReadAll(response.Body); err != nil {
					panic(err)
				} else if response.StatusCode != http.StatusOK {
					panic(fmt.Sprintf(`Expected http status 200, got %d`, response.StatusCode))
				} else if `"HELLO"` != string(result) {
					panic(`Expected result as application/json "HELLO", got ` + string(result))
				} else {
					hs := response.Header[http.CanonicalHeaderKey("Content-Type")]
					if len(hs) != 1 {
						panic("No Content-Type set on response")
					}
					mediaType, _, err := mime.ParseMediaType(hs[0])
					if err != nil {
						panic(fmt.Sprintf("Error parsing content-type: %v", err))
					} else if mediaType != "application/json" {
						panic(fmt.Sprintf("Expected response Content-Type to be set to application/json, got %v", mediaType))
					}
				}
			},
		},
		{
			Name:        "rr-0003",
			Description: "SHOULD reply with 415 on unrecognized Content-Type",
			Optional:    true,
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader("hello"))
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "bogus/content-type")
				req.Header.Set("Accept", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if response.StatusCode != http.StatusUnsupportedMediaType {
					panic(fmt.Sprintf("Expected 415 http code, got %d", response.StatusCode))
				}
			},
		},
		{
			Name:        "rr-0004",
			Description: "MUST reply with 5xx on unmarshalling error",
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader(`"hello`)) // malformed json
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept", "text/plain")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if response.StatusCode < 500 {
					panic(fmt.Sprintf("Expected 5xx http code, got %d", response.StatusCode))
				}
			},
		},
		{
			Name:        "rr-0005",
			Description: "SHOULD reply with 406 on inability to marshall back",
			Optional:    true,
			Image:       "upper",
			T: func(port int) {
				req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/", port), strings.NewReader(`hello`)) // malformed json
				if err != nil {
					panic(err)
				}
				req.Header.Set("Content-Type", "text/plain")
				req.Header.Set("Accept", "not/gonna-happen")
				response, err := http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
				if response.StatusCode != http.StatusNotAcceptable {
					panic(fmt.Sprintf("Expected 406 http code, got %d", response.StatusCode))
				}
			},
		},
	},
}

func setUpContainerUsingPortEnvVar(image string, r *Runner) (*Container, error) {
	_, err := r.dockerClient.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	hostPort, err := getFreePort()
	if err != nil {
		return nil, err
	}
	hostBinding := nat.PortBinding{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", hostPort)}
	containerPort := nat.Port("4321")
	if err != nil {
		return nil, err
	}
	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	cont, err := r.dockerClient.ContainerCreate(context.Background(),
		&container.Config{Image: image, ExposedPorts: nat.PortSet{containerPort: struct{}{}}, Env: []string{"PORT=4321"}},
		&container.HostConfig{
			PortBindings: portBinding,
		}, nil, "")
	if err != nil {
		return nil, err
	}

	err = r.dockerClient.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, err
	}
	for i := 0; i < 10; i = i + 1 {
		_, err = net.Dial("tcp", fmt.Sprintf("localhost:%d", hostPort))
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond * time.Duration(math.Pow(2, float64(i))))
	}
	if err != nil {
		return nil, err
	}
	// TODO: need to sleep some more. Find a more reliable way to diagnose a container as ready
	time.Sleep(1000 * time.Millisecond)

	return &Container{id: cont.ID, hostPort: hostPort}, nil
}
