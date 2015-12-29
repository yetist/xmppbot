#!/bin/bash
go build -o xmppbot.amd64
GOARCH=386 GOOS=linux CGO_ENABLED=1 go build -i -o xmppbot.i386
