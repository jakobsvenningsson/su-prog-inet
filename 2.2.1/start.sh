#!/bin/bash

go run main.go 5000 5001,5002 & P1=$!
go run main.go 5001 5000,5002 & P2=$1
go run main.go 5002 5000,5001

kill ${P1} ${P2}


