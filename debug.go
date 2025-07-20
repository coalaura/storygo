package main

import (
	"encoding/json"
	"os"
)

func dump(v any) {
	b, _ := json.MarshalIndent(v, "", "\t")
	os.WriteFile("dump.json", b, 0755)
}
