package ftp_cmd

type MODE int

const (
	ACTIVE MODE = iota
	PASSIVE
	NOT_SET
)
