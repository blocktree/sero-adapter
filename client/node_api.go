/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package client

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"github.com/blocktree/openwallet/log"
)

type ClientInterface interface {
	Call(path string, request []interface{}) (*gjson.Result, error)
}

// A Client is a Bitcoin RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL     string
	Debug       bool
	Client      *req.Req
}

func NewClient(url string, debug bool) *Client {
	c := Client{
		BaseURL:     url,
		Debug:       debug,
	}

	api := req.New()
	c.Client = api

	return &c
}

// Call calls a remote procedure on another node, specified by the path.
func (c *Client) Call(path string, request []interface{}) (*gjson.Result, error) {

	var (
		body = make(map[string]interface{}, 0)
	)

	if c.Client == nil {
		return nil, errors.New("API url is not setup. ")
	}

	authHeader := req.Header{
		"Accept":        "application/json",
	}

	//json-rpc
	body["jsonrpc"] = "2.0"
	body["id"] = "1"
	body["method"] = path
	body["params"] = request

	if c.Debug {
		log.Std.Info("Start Request API...")
	}

	r, err := c.Client.Post(c.BaseURL, req.BodyJSON(&body), authHeader)

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return nil, err
	}

	result := resp.Get("result")

	return &result, nil
}

// See 2 (end of page 4) http://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

//isError 是否报错
func isError(result *gjson.Result) error {
	var (
		err error
	)

	/*
		//failed 返回错误
		{
			"result": null,
			"error": {
				"code": -8,
				"message": "Block height out of range"
			},
			"id": "foo"
		}
	*/

	if !result.Get("error").IsObject() {

		if !result.Get("result").Exists() {
			return errors.New("Response is empty! ")
		}

		return nil
	}



	errInfo := fmt.Sprintf("[%d]%s",
		result.Get("error.code").Int(),
		result.Get("error.message").String())
	err = errors.New(errInfo)

	return err
}


func (c *Client) LocalSeed2Sk(seed string) (string, error) {
	request := []interface{}{
		seed,
	}

	result, err := c.Call("local_seed2Sk", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (c *Client) LocalSk2Tk(sk string) (string, error) {
	request := []interface{}{
		sk,
	}

	result, err := c.Call("local_sk2Tk", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (c *Client) LocalTk2Pk(tk string) (string, error) {
	request := []interface{}{
		tk,
	}

	result, err := c.Call("local_tk2Pk", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (c *Client) LocalPk2Pkr(pk, rnd string) (string, error) {
	request := []interface{}{
		pk,
		rnd,
	}

	result, err := c.Call("local_pk2Pkr", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}