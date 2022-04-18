package remote

import (
	"strings"
)

type File struct {
	Src string
	Dst string
}

type Cmd struct {
	Cmds   []string
	FileUp []File
}

func (c Cmd) String() string {
	return strings.Join(c.Cmds, " && ")
}

func (c Cmd) List() []string {
	return c.Cmds
}
