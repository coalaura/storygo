package main

import (
	"encoding/json"
	"os"
	"strings"
)

type collector struct {
	b strings.Builder
}

func (c *collector) Write(str string) {
	if !Debug {
		return
	}

	c.b.WriteString(str)
}

func (c *collector) Print(format string) {
	if !Debug {
		return
	}

	defer c.b.Reset()

	debugf(format, c.b.String())
}

func debugd(name string, v any) {
	if !Debug {
		return
	}

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		b = []byte("error: " + err.Error())
	}

	os.WriteFile("_debug-"+name+".json", b, 0755)
}

func debugf(format string, args ...any) {
	if !Debug {
		return
	}

	log.Printf(format+"\n", args...)
}
