package main

import "net"

type Bro struct {
	addr net.Addr
	name string
}

type Bros map[string]Bro

func NewBros() Bros {
	return make(Bros)
}

func (bros Bros) isRoomFull(addr net.Addr) bool {
	if len(bros) < 2 {
		return false
	}
	for k := range bros {
		if k == addr.String() {
			return false
		}
	}
	return true
}

func (bros Bros) otherBro(addr net.Addr) (net.Addr, bool) {
	for k, v := range bros {
		if k != addr.String() {
			return v.addr, true
		}
	}
	return nil, false
}

func (bros Bros) remove(addr net.Addr) {
	delete(bros, addr.String())
}

func (bros Bros) add(addr net.Addr, name string) {
	bros[addr.String()] = Bro{addr: addr, name: name}
}
