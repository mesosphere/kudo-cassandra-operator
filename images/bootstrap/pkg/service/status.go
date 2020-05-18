package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	datacenterPat = regexp.MustCompile(`^\s*Datacenter: (.+)$`)
	nodePat       = regexp.MustCompile(`^\s*([UD][NLJM])\s+([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+)\s+([0-9]+\.[0-9]+ [PGMK]iB)\s+([0-9]+)\s+([0-9\?\.\%]+)\s+([a-zA-Z0-9\-]+)\s+(.+)$`)
)

type Node struct {
	State   string
	Address string
	Load    string
	Tokens  string
	Owns    string
	HostID  string
	Rack    string
}

type Datacenter struct {
	Name  string
	Nodes []Node
}

type Status struct {
	Datacenters []Datacenter
}

func ParseNodetoolStatus(rawStatus string) *Status {
	datacenters := make([]Datacenter, 0)
	for _, line := range strings.Split(rawStatus, "\n") {

		if dcParts := datacenterPat.FindAllStringSubmatch(line, 2); dcParts != nil {
			if len(dcParts[0]) != 2 {
				continue
			}
			datacenters = append(datacenters, Datacenter{Name: dcParts[0][1], Nodes: make([]Node, 0)})
			continue
		}
		if len(datacenters) < 1 {
			continue //without a DC we can't do much else
		}
		if nodeParts := nodePat.FindAllStringSubmatch(line, 7); nodeParts != nil {
			if len(nodeParts[0]) != 8 {
				continue
			}
			datacenters[len(datacenters)-1].Nodes = append(datacenters[len(datacenters)-1].Nodes, Node{State: nodeParts[0][1], Address: nodeParts[0][2], Load: nodeParts[0][3], Tokens: nodeParts[0][4], Owns: nodeParts[0][5], HostID: nodeParts[0][6], Rack: nodeParts[0][7]})
			continue
		}
	}
	return &Status{Datacenters: datacenters}
}

type Nodetool interface {
	Status() (*Status, error)
	Drain() error
	StopDaemon() error
}

type nodetool struct {
	Nodetool string
	Host     string
	Port     string
	SSL      bool
	User     string
	PWF      string
}

// New returns nodetool instance.
func NewNodetool() Nodetool {
	return &nodetool{}
}

func NewRemoteNodetool(ip string, port string) Nodetool {
	userfile := "/etc/cassandra/authentication/username"
	pwfile := "/etc/cassandra/authentication/password"

	user := ""
	pwf := ""
	ssl := false
	if fileExists(userfile) && fileExists(pwfile) {
		ssl = true
		log.Infof("User & PW file exist, assuming SSL connection")
		user, err := ioutil.ReadFile(userfile)
		if err != nil {
			log.Fatalf("failed to read user file: %v", err)
		}
		pw, err := ioutil.ReadFile(pwfile)
		if err != nil {
			log.Fatalf("failed to read pw file: %v", err)
		}
		tmpfile, err := ioutil.TempFile("", "pwf")
		if err != nil {
			log.Fatalf("failed to create temp pw file: %v", err)
		}
		pwFileContent := fmt.Sprintf("%s %s", user, pw)
		if _, err = tmpfile.Write([]byte(pwFileContent)); err != nil {
			log.Fatalf("failed to write temp pw file: %v", err)
		}
		pwf = tmpfile.Name()
		log.Infof("Using User '%s' and PW file '%s'", user, pwf)
	}

	return &nodetool{
		Host: ip,
		Port: port,
		User: user,
		PWF:  pwf,
		SSL:  ssl,
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (n *nodetool) Drain() error {
	cmd := exec.Command("nodetool", n.nodeToolArgs("drain")...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	log.Infof("Drain output:\n%s", out)
	if err != nil {
		return err
	}
	return nil
}

func (n *nodetool) StopDaemon() error {
	cmd := exec.Command("nodetool", n.nodeToolArgs("stopdaemon")...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	log.Infof("StopDaemon output:\n%s", out)
	if err != nil {
		return err
	}
	return nil
}

func (n *nodetool) Status() (*Status, error) {
	cmd := exec.Command("nodetool", n.nodeToolArgs("status")...)
	cmd.Env = os.Environ()
	log.Infof("Run '%s'", cmd.String())
	data, err := cmd.CombinedOutput()
	log.Infof("Status output:\n%s", data)
	if err != nil {
		return nil, err
	}
	return ParseNodetoolStatus(string(data)), nil
}

func (n *nodetool) nodeToolArgs(additionalArgs ...string) []string {
	args := []string{}
	if n.Host != "" {
		args = append(args, "-h", n.Host)
	}
	if n.Port != "" {
		args = append(args, "-p", n.Port)
	}
	if n.User != "" {
		args = append(args, "-u", n.User)
	}
	if n.PWF != "" {
		args = append(args, "-pwf", n.PWF)
	}
	if n.SSL {
		args = append(args, "--ssl")
	}

	args = append(args, additionalArgs...)
	log.Infof("Nodetool args: %v", args)
	return args
}
