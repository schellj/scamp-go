package scamp

type msgNoType int64

type Message interface {
	toPackets(msgNoType) []Packet
}