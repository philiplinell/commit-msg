/*
	Package build provides a way to get information about the current build.

	Note that you need to build with the full package path to correctly insert the build information.

	E.g.:
		go build -o bin/helper github.com/philiplinell/commit-msg/cmd/cli
	instead of
		go build -o bin/helper ./cmd/cli/*.go

	See more here: https://github.com/golang/go/issues/51831
*/
package build

import (
	"errors"
	"fmt"
	"runtime/debug"
	"time"
)

type Info struct {
	// Commit - the commit identifier for the current build.
	Commit string

	// Dirty - true if the source tree had local modifications.
	Dirty bool

	// ModificationTime - the modification time associated with the commit.
	ModificationTime time.Time
}

const (
	// vcsRevisionKey - revision identifier for the current commit or checkout
	vcsRevisionKey = "vcs.revision"

	// vscTimeKey - the modification time associated with vcs.revision, in RFC3339 format
	vcsTimeKey = "vcs.time"

	// vcsModifiedKey - true or false indicating whether the source tree had local modifications
	vcsModifiedKey = "vcs.modified"
)

// GetInfo - returns the build information.
func GetInfo() (*Info, error) {
	debugInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return nil, errors.New("could not read build info")
	}

	info := &Info{}

	for _, setting := range debugInfo.Settings {
		if setting.Key == vcsRevisionKey {
			info.Commit = setting.Value
		}
		if setting.Key == vcsTimeKey {
			t, err := time.Parse(time.RFC3339, setting.Value)
			if err != nil {
				return nil, fmt.Errorf("could not parse time value: %w", err)
			}

			info.ModificationTime = t
		}

		if setting.Key == vcsModifiedKey {
			info.Dirty = setting.Value == "true"
		}
	}

	return info, nil
}

// String - returns a string representation of the build info.
func (i *Info) String() string {
	if i.Dirty {
		return fmt.Sprintf("%s-dirty %s", i.Commit, i.ModificationTime)
	}

	return fmt.Sprintf("%s %s", i.Commit, i.ModificationTime)
}
