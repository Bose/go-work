#!/bin/bash
go test -v -covermode=count -coverprofile=count.out  && go tool cover -html=count.out
