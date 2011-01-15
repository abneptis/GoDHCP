package dhcp

import "net"
import "os"
import "sync"
import "time"

type AddressFinderRequest struct {
  Timeout        int64
  MaximumOffers  int
  Options        []Option
  IfType         uint8
  IfAddr         []byte
}

type AddressFinderResponse []Message

func (self AddressFinderResponse)Len()(int){return len(self)}

type AddressAcceptRequest struct {
  Timeout int64
  Offer   Message
}

// This is currently focused around DHCP,
// but presumably could be extended to other protocols
// presuming we instead had a more generic 'lease' object.
type AddressFinder interface {
  DiscoverOffers(*AddressFinderRequest, *AddressFinderResponse)(os.Error)
  AcceptOffer(*AddressAcceptRequest, *Message)(os.Error)
}


type dhcpFinder struct {
  ifname string
  sock *dhcp4Socket
  socklock sync.Mutex
}

func NewDHCP4Finder(ifname string)(af AddressFinder, err os.Error){
 if err == nil {
   _af := &dhcpFinder{ifname: ifname}
   _af.sock, err = newDhcp4Socket(ifname)
   if err == nil {
     af = _af
   }
 }
 return
}

func (self *dhcpFinder)DiscoverOffers(req *AddressFinderRequest, resp *AddressFinderResponse)(err os.Error){
  *resp = AddressFinderResponse{}
  msg, err := NewMessage(0, BOOT_REQUEST, req.IfType, req.IfAddr, nil, nil, nil, nil)
  if err == nil {
    msg.SetDHCPMessageType(DHCPDISCOVER)
    start := time.Nanoseconds()
    self.socklock.Lock()
    defer self.socklock.Unlock()
    if req.Timeout <= 0 {
      // default to 15 Seconds.
      req.Timeout = 15*1000000000
    }
    err = self.sock.sock.SetWriteTimeout(req.Timeout)
    if err != nil { return }
    err = self.sock.WriteMessage(msg, net.IPv4bcast)
    if err != nil { return }
    for {
      var omsg *Message
      to := req.Timeout - (time.Nanoseconds() - start)
      if to < 0 { return }
      err = self.sock.sock.SetReadTimeout(to)
      if err != nil { return }
      omsg, _, err = self.sock.ReadMessage()
      if err == nil {
        if omsg.Xid != msg.Xid {
          //log.Printf("Warning: Skipping unexpected DHCP (XID mismatch)")
          continue
        }
        *resp = append(*resp, *omsg)
        if req.MaximumOffers > 0 && resp.Len() >= req.MaximumOffers {
          return
        }
      }
    }
  }
  // handle timeout...
  return
}

func (self *dhcpFinder)AcceptOffer(req *AddressAcceptRequest, resp *Message)(err os.Error){
  req.Offer.Operation = BOOT_REQUEST
  req.Offer.Options = []Option{
    NewOption(DHCPOPT_REQUEST_IP,[]byte(req.Offer.YourIP.To4())),
    NewOption(DHCPOPT_SERVER_ID,[]byte(req.Offer.ServerIP.To4())),
  }
  req.Offer.SetDHCPMessageType(DHCPREQUEST)
  req.Offer.YourIP = net.IPv4(0,0,0,0)
  start := time.Nanoseconds()
  self.socklock.Lock()
  defer self.socklock.Unlock()
  if req.Timeout <= 0 {
    // default to 15 Seconds.
    req.Timeout = 15*1000000000
  }
  err = self.sock.sock.SetWriteTimeout(req.Timeout)
  if err != nil { return }
  err = self.sock.WriteMessage(&req.Offer, net.IPv4bcast)
  if err != nil { return }
  to := req.Timeout - (time.Nanoseconds() - start)
  if to < 0 { return }
  err = self.sock.sock.SetReadTimeout(to)
  if err != nil { return }
  var msg *Message
  msg, _, err = self.sock.ReadMessage()
  if err == nil {
    *resp = *msg
  }
  return
}
