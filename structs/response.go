package structs

import (
	"encoding/json"
	"fmt"
)

// Response is the JSONL envelope understood by Docker Compose / the provider protocol.
type Response struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (r *Response) PrintAsString() {
	b, _ := json.Marshal(r)
	fmt.Println(string(b))
}
