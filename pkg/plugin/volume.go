package plugin

import (
	"fmt"
	"strings"
)

type MountVolume struct {
	Source string
	Target string
	FsType string
	Options []string
}

func (m MountVolume) String() string {
	return fmt.Sprintf("%s %s %s %s", m.Source, m.Target, m.FsType, strings.Join(m.Options, ","))
}
