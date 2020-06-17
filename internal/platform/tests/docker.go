package tests

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"
)

// Container tracks information about a docker container started for tests.
type Container struct {
	ID   string
	Host string // IP:Port
}

// startContainer runs a postgres container to execute commands.
func startContainer(t *testing.T, image string) *Container {
	t.Helper()

	cmd := exec.Command("docker", "run", "-P", "-d", "-e", "POSTGRES_PASSWORD=postgres", image)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container %s: %v", image, err)
	}

	id := out.String()[:12]

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
	}

	var doc []struct {
		NetworkSettings struct {
			Ports struct {
				TCP5432 []struct {
					HostIP   string `json:"HostIp"`
					HostPort string `json:"HostPort"`
				} `json:"5432/tcp"`
			} `json:"Ports"`
		} `json:"NetworkSettings"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}

	dbHost := doc[0].NetworkSettings.Ports.TCP5432[0]

	c := Container{
		ID:   id,
		Host: dbHost.HostIP + ":" + dbHost.HostPort,
	}

	t.Logf("DB ContainerID: %s", c.ID)
	t.Logf("DB Host: %s", c.Host)

	return &c
}

// stopContainer stops and removes the specified container.
func stopContainer(t *testing.T, id string) {
	t.Helper()

	if err := exec.Command("docker", "stop", id).Run(); err != nil {
		t.Fatalf("could not stop container: %v", err)
	}
	t.Log("Stopped:", id)

	if err := exec.Command("docker", "rm", id, "-v").Run(); err != nil {
		t.Fatalf("could not remove container: %v", err)
	}
	t.Log("Removed:", id)
}

// dumpContainerLogs runs "docker logs" against the container and send it to t.Log
func dumpContainerLogs(t *testing.T, id string) {
	t.Helper()

	out, err := exec.Command("docker", "logs", id).CombinedOutput()
	if err != nil {
		t.Fatalf("could not log container: %v", err)
	}
	t.Logf("Logs for %s\n%s:", id, out)
}
