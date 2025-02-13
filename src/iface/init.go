package iface

import (
	"fmt"
	"github.com/shirou/gopsutil/v4/net"
)

var Iface map[string]net.InterfaceStat

func init() {
	ifaces, err := net.Interfaces()
	if err != nil {
		panic(fmt.Sprintf("Error getting interfaces: %s", err.Error()))
	}

	Iface = make(map[string]net.InterfaceStat, len(ifaces))

	for _, iface := range ifaces {
		Iface[iface.Name] = iface
	}
}
