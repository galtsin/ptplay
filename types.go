package ptplay

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Request struct {
	A int `json:"a"`
	B int `json:"b"`
}

type Response struct {
	S int `json:"s"`
	M int `json:"m"`
}

type Result struct {
	Request
	Response
}

func (r Request) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Request) Unmarshal(bs []byte) error {
	return json.Unmarshal(bs, r)
}

func (r Response) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Response) Unmarshal(bs []byte) error {
	return json.Unmarshal(bs, r)
}

func (r Result) Marshal() ([]byte, error) {
	return bytes.NewBufferString(fmt.Sprintf("a:%d,b:%d,s:%d,m:%d", r.A, r.B, r.S, r.M)).Bytes(), nil
}
