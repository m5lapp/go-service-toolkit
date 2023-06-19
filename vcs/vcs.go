package vcs

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	var timestamp string
	var revision string
	var modified bool

	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.time":
				timestamp = s.Value
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					modified = true
				}
			}
		}
	}

	if modified {
		return fmt.Sprintf("%s-%s-dirty", timestamp, revision)
	}

	return fmt.Sprintf("%s-%s", timestamp, revision)
}
