package logging

import (
	"net"

	"go.uber.org/zap/zapcore"
)

type IPs []net.IP

func (a IPs) MarshalLogArray(enc zapcore.ArrayEncoder) error {
	for _, ip := range a {
		enc.AppendString(ip.String())
	}
	return nil
}
