package debug

import (
	"encoding/json"
	"fmt"
)

var DebugOutput bool = false

func Log(format string, args ...any) {
	if !DebugOutput {
		return
	}
	fmt.Printf(format, args...)
}

func LogJSON(prefix string, object any, suffix string) {
	if !DebugOutput {
		return
	}

	json, _ := json.MarshalIndent(object, "", "    ")
	fmt.Printf("%s%s%s", prefix, json, suffix)
}
