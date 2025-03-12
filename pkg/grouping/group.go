package grouping

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/akerouanton/gostackdrill/pkg/utils"
)

type GroupByType string

const (
	GroupByPackage GroupByType = "package"
	GroupByStruct  GroupByType = "struct"
	GroupByFunc    GroupByType = "func"
)

type Group struct {
	Key         string
	Goroutines  []*gostackparse.Goroutine
	CumWaitTime time.Duration
}

func (g Group) AvgWaitTime() time.Duration {
	return time.Duration(int64(g.CumWaitTime) / int64(len(g.Goroutines)))
}

func GroupGoroutines(goroutines []*gostackparse.Goroutine, groupBy GroupByType, frameFilters ...func(*gostackparse.Frame) bool) map[string]Group {
	groups := make(map[string]Group)
	for _, goroutine := range goroutines {
		frame, _ := utils.FindFirstFrame(goroutine.Stack, frameFilters...)
		if frame == nil {
			continue
		}

		key := ""
		if groupBy == GroupByPackage {
			key = filepath.Dir(frame.File)
		} else if groupBy == GroupByStruct {
			key = structKey(frame)
		} else if groupBy == GroupByFunc {
			key = frame.Func
		} else {
			panic(fmt.Sprintf("unknown group by: %s", groupBy))
		}

		group, ok := groups[key]
		if !ok {
			group = Group{
				Key:         key,
				Goroutines:  make([]*gostackparse.Goroutine, 0),
				CumWaitTime: 0,
			}
		}
		group.Goroutines = append(group.Goroutines, goroutine)
		group.CumWaitTime += goroutine.Wait
		groups[key] = group
	}

	return groups
}

func structKey(frame *gostackparse.Frame) string {
	lastSlash := strings.LastIndex(frame.Func, "/")
	firstDot := strings.Index(frame.Func[lastSlash+1:], ".")
	return frame.Func[:lastSlash+1+firstDot]
}
