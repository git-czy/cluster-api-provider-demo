package remote

import (
	"strings"
)

type File struct {
	Src string `json:"src,omitempty"`
	Dst string `json:"dst,omitempty"`
}

type Cmd struct {
	Cmds   []string `json:"cmds,omitempty"`
	FileUp []File   `json:"fileUp,omitempty"`
}

func (c Cmd) String() string {
	return strings.Join(c.Cmds, " && ")
}

func (c Cmd) List() []string {
	return c.Cmds
}

func (in *Cmd) DeepCopyInto(out *Cmd) {
	*out = *in
}

func (in *Cmd) DeepCopy() *Cmd {
	if in == nil {
		return nil
	}
	out := new(Cmd)
	in.DeepCopyInto(out)
	return out
}
