package dhcp

import "net"
import "os"

type dhcp4Socket struct {
  sock *net.UDPConn
}



// dhcp.Bind("eth0", "0.0.0.0:68")
func bind(ifname, bindaddr string)(c *net.UDPConn, err os.Error){
  laddr, err := net.ResolveUDPAddr(bindaddr)
  if err != nil { return }
  c,  err = net.ListenUDP("udp4",laddr)
  if err != nil { return }
  err = c.BindToDevice(ifname)
  if err != nil {
    c.Close()
    c = nil
  }
  return
}

func newDhcp4Socket(ifname string)(c *dhcp4Socket, err os.Error){
  c = &dhcp4Socket{}
  if err == nil {
    c.sock, err = bind(ifname, "0.0.0.0:68")
  }
  return
}

func (self *dhcp4Socket)WriteMessage(msg *Message, src net.IP)(err os.Error){
  err = WriteMessage(self.sock, msg, src)
  return
}
func (self *dhcp4Socket)ReadMessage()(*Message, net.Addr, os.Error){
  return ReadMessage(self.sock)
}

func (self *dhcp4Socket)Close()(os.Error){
  return self.sock.Close()
}
