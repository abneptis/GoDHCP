package dhcp

import "log"
import "os"
import "net"
import "bytes"

type Client struct {
  sock *SocketV4
  ifname string
  ifhwtype byte
  ifhwaddr []byte
  xidChannels map[uint32]chan Message
  running bool
  offerSelector OfferSelector
  offerAcceptor OfferAcceptor
  offers []Message
}

func NewClient(ifname string, mactype byte, mac []byte)(c *Client, err os.Error){
  c = &Client {
    ifname: ifname,
    ifhwtype: mactype,
    ifhwaddr: mac,
    offerSelector: fifoSelector{},
    offerAcceptor: defaultAcceptor{}, 
  }
  c.sock, err = NewSocketV4(ifname, mactype, mac)
  return
}

func (self *Client)read()(err os.Error){
  for self.running {
    msg, _, err := self.sock.ReadMessage()
    if err != nil {
      log.Printf("Critical: Socket error: %v", err)
      continue
    }
    t, err := msg.DHCPMessageType()
    if err != nil {
      log.Printf("Warning: Received a message with no DHCP Type: %v", err)
      continue
    }
    if 0 != bytes.Compare(self.ifhwaddr, msg.ClientHWAddr[0:len(self.ifhwaddr)]){
      log.Printf("Info:  Received a message not targeted to me %v", msg)
      continue
    }
    switch (t) {
      case DHCPACK: err = self.onAck(*msg)
      case DHCPOFFER: self.onOffer(*msg)
    }
    if err != nil {
      log.Printf("Critical: Upstream handler error: %v", err)
    }
  }
  return
}

func (self *Client)onAck(msg Message)(err os.Error){
  err = self.offerAcceptor.AcceptOffer(msg)
  if err != nil {
    log.Printf("Acceptor rejected message (%v);  Should NAK", err)
  }
  return
}

func (self *Client)onOffer(msg Message){
  self.offers = append(self.offers, msg)
  o, err := self.offerSelector.SelectOffer(self.offers)
  if err == nil {
    err = self.Request(o)
    if err != nil {
      log.Printf("Error: requesting offer: %v", err)
    }
  }
}

func (self *Client)StopReader(){
  self.running = false
}

func (self *Client)StartReader(){
  self.running = true
  go self.read()
}

func (self *Client)Discover(opts []Option)(err os.Error){
  msg, err := NewMessage(0, BOOT_REQUEST, self.ifhwtype, self.ifhwaddr, nil, nil, nil, nil)
  if err == nil {
    opts = append(opts, NewOption(DHCPOPT_MESSAGE_TYPE, []byte{DHCPDISCOVER}))
    msg.Options = opts
    err = self.sock.WriteMessage(msg, net.IPv4bcast)
  }
  return
}

func (self *Client)Request(offer Message)(err os.Error){
  t, err := offer.DHCPMessageType()
  if err != nil { return }
  if t != DHCPOFFER {
    err = os.NewError("Cannot request an unoffered address")
    return
  }
  request := &offer
  request.Operation = BOOT_REQUEST
  ipopt, _ := IP4Option(DHCPOPT_REQUEST_IP, offer.YourIP)
  svropt, _ := IP4Option(DHCPOPT_SERVER_ID, offer.ServerIP)
  request.Options = []Option{ipopt, svropt, NewOption(DHCPOPT_MESSAGE_TYPE, []byte {DHCPREQUEST})}
  request.YourIP = net.IPv4(0,0,0,0)
  if err == nil {
    err = self.sock.WriteMessage(request, net.IPv4bcast)
  }
  return
}


