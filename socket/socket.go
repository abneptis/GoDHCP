package dhcp

import "net"
import "os"

type SocketV4 struct {
  sock *net.UDPConn
  hardwareType byte
  hardwareAddr [16]byte
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

func NewSocketV4(ifname string, ht byte , ha []byte)(c *SocketV4, err os.Error){
  c = &SocketV4{ hardwareType: ht }
  switch ht {
    case ETHERNET:
      copy(c.hardwareAddr[0:6], ha[0:6])
    default:
      err = os.NewError("Can't handle hardware type -- care to send a patch?")
  }
  if err == nil {
    c.sock, err = bind(ifname, "0.0.0.0:68")
  }
  return
}

func (self *SocketV4)WriteMessage(msg *Message, src net.IP)(err os.Error){
  err = WriteMessage(self.sock, msg, src)
  return
}
func (self *SocketV4)ReadMessage()(*Message, net.Addr, os.Error){
  return ReadMessage(self.sock)
}

