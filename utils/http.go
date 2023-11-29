package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func DoHttpRequest(args RequestArgs) (any, int, error) {
	b, err := json.Marshal(args.Data)
	if err != nil {
		return nil, 0, err
	}

	body := bytes.NewReader(b)

	req, err := http.NewRequest(args.Method, args.Url, body)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Add("content-type", "application/json")
	req.SetBasicAuth(args.Username, args.Password)

	c := http.Client{}

	resp, err := c.Do(req)
	if err != nil {
		return nil, 0, err
	}

	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	r := bytes.NewReader(b)
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	var v any
	err = decoder.Decode(&v)
	if err != nil {
		return nil, 0, err
	}

	return v, resp.StatusCode, nil
}

type RequestArgs struct {
	Method             string
	Url                string
	Data               any
	Username, Password string
}
