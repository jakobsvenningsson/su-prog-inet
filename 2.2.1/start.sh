#!/bin/bash

go run main.go 5000 127.0.0.1 5001 & P1=$!
go run main.go 5001 127.0.0.1 5000

kill ${P1}


