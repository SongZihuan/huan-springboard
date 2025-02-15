package sshserver

import (
	"github.com/SongZihuan/huan-springboard/src/config"
	"net"
)

type SshController interface {
	RemoteAddrCheck(remoteAddr *net.TCPAddr, to *net.TCPAddr, countRules []*config.SshCountRuleConfig) error
}
