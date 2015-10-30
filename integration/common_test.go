package integration

import (
	"bytes"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/docker"
	"github.com/samalba/dockerclient"

	. "gopkg.in/check.v1"
)

const SimpleTemplate = `
        hello:
          image: busybox
          stdin_open: true
          tty: true
        `

func Test(t *testing.T) { TestingT(t) }

func asMap(items []string) map[string]bool {
	result := map[string]bool{}
	for _, item := range items {
		result[item] = true
	}
	return result
}

var random = rand.New(rand.NewSource(time.Now().Unix()))

func RandStr(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}
	return string(b)
}

type RunSuite struct {
	command  string
	projects []string
}

func (s *RunSuite) TearDownTest(c *C) {
	// Delete all containers
	client := GetClient(c)
	containers, err := client.ListContainers(true, false, "")
	c.Assert(err, IsNil)
	for _, container := range containers {
		err := client.RemoveContainer(container.Id, true, true)
		c.Assert(err, IsNil)
	}
}

var _ = Suite(&RunSuite{
	command: "../bundles/libcompose-cli_linux-amd64",
})

func (s *RunSuite) CreateProjectFromText(c *C, input string) string {
	return s.ProjectFromText(c, "create", input)
}

func (s *RunSuite) RandomProject() string {
	return "test-project-" + RandStr(7)
}

func (s *RunSuite) ProjectFromText(c *C, command, input string) string {
	projectName := s.RandomProject()
	return s.FromText(c, projectName, command, input)
}

func (s *RunSuite) FromText(c *C, projectName, command string, argsAndInput ...string) string {
	args := []string{"--verbose", "-p", projectName, "-f", "-", command}
	args = append(args, argsAndInput[0:len(argsAndInput)-1]...)

	input := argsAndInput[len(argsAndInput)-1]

	if command == "up" {
		args = append(args, "-d")
	} else if command == "down" {
		args = append(args, "--timeout", "0")
	} else if command == "restart" {
		args = append(args, "--timeout", "0")
	} else if command == "stop" {
		args = append(args, "--timeout", "0")
	}

	logrus.Infof("Running %s %v", command, args)

	cmd := exec.Command(s.command, args...)
	cmd.Stdin = bytes.NewBufferString(strings.Replace(input, "\t", "  ", -1))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Failed to run %s %v: %v\n with input:\n%s", s.command, err, args, input)
	}

	c.Assert(err, IsNil)

	return projectName
}

func GetClient(c *C) dockerclient.Client {
	client, err := docker.CreateClient(docker.ClientOpts{})

	c.Assert(err, IsNil)

	return client
}

func (s *RunSuite) GetContainerByName(c *C, name string) *dockerclient.ContainerInfo {
	client := GetClient(c)
	container, err := docker.GetContainerByName(client, name)

	c.Assert(err, IsNil)

	if container == nil {
		return nil
	}

	info, err := client.InspectContainer(container.Id)

	c.Assert(err, IsNil)

	return info
}

func (s *RunSuite) GetContainersByProject(c *C, project string) []dockerclient.Container {
	client := GetClient(c)
	containers, err := docker.GetContainersByFilter(client, docker.PROJECT.Eq(project))

	c.Assert(err, IsNil)

	return containers
}
