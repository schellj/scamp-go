// Copyright 2014 GüdTech, Inc.

/*
SCAMP provides SOA bus RPC functionality. Please see root SCAMP/README.md for details on configuring environment.

Basics

Services and requesters communicate over persistent TLS connections. First, initialize your environment according to the root README.md. You must have a valid certificate and key to present a service.

Listening For Requests

  service,err := scamp.NewService(30100)
  if err != nil {
	fmt.Printf("could not create service")
	return
  }
  service.Register("helloworld.hello", func(sess *scamp.Session){
  })
  service.AcceptSessions()

Making a Request against a Service

  conn,err := scamp.Connect(":30100")
  if err != nil {
	fmt.Printf("could not connect to service")
	return
  }
  sess,err := conn.Send(scamp.Request{
    Action:         "helloworld.hello",
    EnvelopeFormat: scamp.ENVELOPE_JSON,
    Version:        1,
  })
  if err != nil {
	fmt.Printf("could not send reqest")
	return
  }
  var reply Reply = sess.Recv()

Library Internals

SCAMP is a layered architecture:

  Request/Reply
  -------------
  Session
  -------------
  Connection
  -------------
  Service

*/
package scamp

import (
  "fmt"
)

var defaultCache *ServiceCache

// Package-level setup. Right now it just sets up logging.
func Initialize(configPath string) (err error) {
	initSCAMPLogger()
  err = initConfig(configPath)
  if err != nil {
    return
  }

  cachePath,found := DefaultConfig().Get("discovery.cache_path")
  if !found {
    err = fmt.Errorf("no such config param `discovery.cache_path`. must be set to use scamp-go.")
    return
  }

  defaultCache,err = NewServiceCache(cachePath)
  if err != nil {
    return
  }

  return
}