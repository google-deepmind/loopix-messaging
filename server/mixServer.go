/*
	Package server implements the mix server.
*/
package server

import (
	"fmt"
	"net"

	"anonymous-messaging/networker"
	"anonymous-messaging/node"
	"anonymous-messaging/config"
	"anonymous-messaging/helpers"
	"anonymous-messaging/sphinx"
	log "github.com/sirupsen/logrus"
	"github.com/protobuf/proto"
)

type MixServerIt interface {
	networker.NetworkServer
	networker.NetworkClient
}

type MixServer struct {
	Id   string
	Host string
	Port string
	Listener *net.TCPListener
	node.Mix

	Config config.MixConfig
}

func (m *MixServer) ReceivedPacket(packet []byte) error{
	log.WithFields(log.Fields{"id" : m.Id}).Info("Received new sphinx packet")

	c := make(chan []byte)
	cAdr := make(chan sphinx.Hop)
	cFlag := make(chan string)
	errCh := make(chan error)

	go m.ProcessPacket(packet, c, cAdr, cFlag, errCh)
	dePacket := <-c
	nextHop := <- cAdr
	flag := <- cFlag
	err := <- errCh

	if err != nil{
		return err
	}

	if flag == "\xF1" {
		m.ForwardPacket(dePacket, nextHop.Address)
	} else  {
		log.WithFields(log.Fields{"id" : m.Id}).Info("Packet has non-forward flag. Packet dropped")
	}
	return nil
}

func (m *MixServer) ForwardPacket(sphinxPacket []byte, address string) error{
	packetBytes, err := config.WrapWithFlag(COMM_FLAG, sphinxPacket)
	if err != nil{
		return err
	}
	err = m.Send(packetBytes, address)
	if err != nil{
		return err
	}

	return nil
}

func (m *MixServer) Send(packet []byte, address string) error{

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		return err
	}
	return nil
}

func (m *MixServer) Start() error {
	defer m.Run()

	return nil
}

func (m *MixServer) Run() {

	defer m.Listener.Close()
	finish := make(chan bool)

	go func() {
		log.WithFields(log.Fields{"id" : m.Id}).Info(fmt.Sprintf("Listening on %s", m.Host + ":" + m.Port))
		m.ListenForIncomingConnections()
	}()

	<-finish
}

func (m *MixServer) ListenForIncomingConnections() {
	for {
		conn, err := m.Listener.Accept()

		if err != nil {
			log.WithFields(log.Fields{"id" : m.Id}).Error(err)
		} else {
			log.WithFields(log.Fields{"id" : m.Id}).Info(fmt.Sprintf("Received connection from %s", conn.RemoteAddr()))
			errs := make(chan error, 1)
			go m.HandleConnection(conn, errs)
			err = <-errs
			if err != nil{
				log.WithFields(log.Fields{"id" : m.Id}).Error(err)
			}
		}
	}
}

func (m *MixServer) HandleConnection(conn net.Conn, errs chan<- error) {
	defer conn.Close()

	buff := make([]byte, 1024)
	reqLen, err := conn.Read(buff)
	if err != nil {
		errs <- err
	}

	var packet config.GeneralPacket
	err = proto.Unmarshal(buff[:reqLen], &packet)
	if err != nil {
		errs <- err
	}

	switch packet.Flag {
	case COMM_FLAG:
		err = m.ReceivedPacket(packet.Data)
		if err != nil{
			errs <- err
		}
	default:
		log.WithFields(log.Fields{"id" : m.Id}).Info("Packet flag not recognised. Packet dropped.")
	}
}

func NewMixServer(id, host, port string, pubKey []byte, prvKey []byte, pkiPath string) (*MixServer, error) {
	node := node.Mix{Id: id, PubKey: pubKey, PrvKey: prvKey}
	mixServer := MixServer{Id: id, Host: host, Port: port, Mix: node, Listener: nil}
	mixServer.Config = config.MixConfig{Id : mixServer.Id, Host: mixServer.Host, Port: mixServer.Port, PubKey: mixServer.PubKey}

	configBytes, err := proto.Marshal(&mixServer.Config)
	if err != nil{
		return nil, err
	}
	err = helpers.AddToDatabase(pkiPath, "Pki", mixServer.Id, "Mix", configBytes)
	if err != nil{
		return nil, err
	}

	addr, err := helpers.ResolveTCPAddress(mixServer.Host, mixServer.Port)

	if err != nil {
		return nil, err
	}
	mixServer.Listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		return nil, err
	}

	return &mixServer, nil
}
