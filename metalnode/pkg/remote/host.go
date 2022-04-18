package remote

import (
	"fmt"
	"net"
)

type Host struct {
	User     string
	Password string
	Address  string
	Port     int
	SSHKey   string
}

func (h *Host) Validate() (*Host, error) {
	if h.User == "" {
		return nil, fmt.Errorf("Host's user field is required ")

	}
	if h.Password == "" && h.SSHKey == "" {
		return nil, fmt.Errorf("At least one of the host's password and ssh key is provided ")
	}
	if h.Address == "" {
		return nil, fmt.Errorf("Host address is required ")
	}
	if a := net.ParseIP(h.Address); a == nil {
		return nil, fmt.Errorf("Host's address not a valid IP address ")
	}
	if h.Port < 0 {
		return nil, fmt.Errorf("Host's port must be greater than zero ")
	}
	return h, nil
}

func (h Host) Fields() (string, string, string, int, string) {
	return h.User, h.Password, h.Address, h.Port, h.SSHKey
}
