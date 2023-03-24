package main

import (
	"bytes"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

type JsVmClient struct {
	runtime *goja.Runtime
	value   goja.Value

	GetRoomsId func() []string
	OnMessage  func(*Message)
}

func (c *JsVmClient) OnMessageSafe(message *Message) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("OnMessage panic: %v", e)
		}
	}()
	if c.OnMessage == nil {
		return nil
	}
	c.OnMessage(message)
	return
}

func (c *JsVmClient) GetRoomsIdSafe() (rooms []string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("GetRoomsId panic: %v", e)
		}
	}()
	if c.GetRoomsId == nil {
		return []string{}, nil
	}
	rooms = c.GetRoomsId()
	return
}

func NewJsVmFromFile(fileName string) (*JsVmClient, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return NewJsVmFromString(string(content))
}

func NewJsVmFromString(content string) (*JsVmClient, error) {
	rt := goja.New()
	value, err := rt.RunString(content)
	if err != nil {
		return nil, err
	}

	jsVm := &JsVmClient{
		runtime: rt,
		value:   value,
	}

	getRoomsId := rt.Get("get_rooms_id")
	if getRoomsId != nil {
		if err := rt.ExportTo(getRoomsId, &jsVm.GetRoomsId); err != nil {
			jsVm.GetRoomsId = nil
		}
	}

	onMessage := rt.Get("on_message")
	if onMessage != nil {
		if err := rt.ExportTo(onMessage, &jsVm.OnMessage); err != nil {
			jsVm.OnMessage = nil
		}
	}

	if err = rt.Set("echo", log.Println); err != nil {
		return nil, err
	}

	if err = rt.Set("bytes2string", bytes2string); err != nil {
		return nil, err
	}

	if err = rt.Set("string2bytes", string2bytes); err != nil {
		return nil, err
	}

	if err = rt.Set("http_request", httpRequest); err != nil {
		return nil, err
	}

	return jsVm, nil
}

func bytes2string(b []byte) string {
	return string(b)
}

func string2bytes(s string) []byte {
	return []byte(s)
}

type request struct {
	Url     string
	Method  string
	Params  map[string][]string
	Body    []byte
	Headers map[string]string
	Proxy   struct {
		Http  *string
		Https *string
	}
}

func httpRequest(req request) (content []byte) {
	u, err := url.Parse(req.Url)
	if err != nil {
		log.Printf("url parse failure: %s (%v)\n", req.Url, err)
		return
	}

	if req.Params != nil {
		q := u.Query()
		for k, vs := range req.Params {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
	}

	method := strings.ToUpper(req.Method)
	if method != "GET" && method != "POST" {
		log.Printf("unsupport request method: %s (only 'GET', 'POST')\n", req.Method)
		return
	}

	var body io.Reader = nil
	if req.Body != nil {
		body = bytes.NewReader(req.Body)
	}

	r, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		log.Printf("new request failure: %v\n", err)
		return
	}

	for k, v := range req.Headers {
		r.Header.Set(k, v)
	}

	cli := http.Client{}

	proxyVal := reflect.ValueOf(req.Proxy)
	proxyTyp := proxyVal.Type()
	for i := 0; i < proxyTyp.NumField(); i++ {
		k := proxyTyp.Field(i).Name
		v := proxyVal.Field(i)
		if v.IsZero() {
			continue
		}

		pro, err := url.Parse(*(v.Interface().(*string)))
		if err != nil {
			fmt.Printf("set %s proxy failure: %s (%v)\n", k, *(v.Interface().(*string)), err)
			continue
		}

		if strings.ToLower(k) == "http" || strings.ToLower(k) == "https" {
			cli.Transport = &http.Transport{
				Proxy: http.ProxyURL(pro),
			}
			break
		}
	}

	resp, err := cli.Do(r)
	if err != nil {
		log.Printf("request failure: %v\n", err)
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	content, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read response failure: %v\n", err)
		return
	}

	return
}
