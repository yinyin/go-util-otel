package utilotel

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func splitHostPort(hostPort string) (host string, port int) {
	portMultiplier := 1
	for idx := len(hostPort) - 1; idx >= 0; idx-- {
		ch := hostPort[idx]
		if ch == ':' {
			host = hostPort[:idx]
			if (len(host) > 2) && (host[0] == '[') && (host[len(host)-1] == ']') {
				host = host[1 : len(host)-1]
			}
			return
		}
		if ch >= '0' && ch <= '9' {
			port += int(ch-'0') * portMultiplier
			portMultiplier *= 10
		} else {
			port = 0
			host = hostPort
			return
		}
	}
	return
}

// AppendNetPeerConnAttributes append network peer attributes for connection.
// The appended attributes conforming to the "network.peer.address" and "network.peer.port"
// semantic conventions.
func AppendNetPeerConnAttributes(attrs []attribute.KeyValue, hostPort string) []attribute.KeyValue {
	addrHost, addrPort := splitHostPort(hostPort)
	if addrHost != "" {
		attrs = append(attrs, semconv.NetworkPeerAddress(addrHost))
	}
	if addrPort != 0 {
		attrs = append(attrs, semconv.NetworkPeerPort(addrPort))
	}
	return attrs
}
