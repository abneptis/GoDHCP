package dhcp

import "rand"
import "net"
import "os"
import "encoding/binary"
//import "log"


type Message struct {
  Operation byte
  HardwareType byte
  HardwareLen uint8
  Hops uint8
  Xid uint32
  Secs uint16
  Flags uint16
  ClientIP net.IP
  YourIP net.IP
  ServerIP net.IP
  GatewayIP net.IP
  ClientHWAddr [16]byte
  // These are odd legacy 'sortofrepurposed' from BOOTP.
  // They are /not/ typically the strings they're named for.
  // The bored could read rfc2131
  Options []Option
}

func NewMessage(xid uint32, op, hwtype byte, hwaddr []byte, C,Y,S,G net.IP)(msg *Message, err os.Error){
  // select a random XID
  if xid == 0 {
    xid = rand.Uint32()
  }
  if len(hwaddr) > 16 {
    err = os.NewError("Invalid DHCP hardware address (too long)")
    return
  }
  msg = &Message {
    Operation: op,
    HardwareType: hwtype,
    HardwareLen: byte(len(hwaddr)),
    Xid: xid,
    // Secs: 0?
    // We default to FLAG_BROADCAST to avoid needing raw packets in Linux
    // This works for EC2, ISC, and possibly others, but no idea where else.
    Flags: FLAG_BROADCAST,
    ClientIP: C,
    YourIP: Y,
    ServerIP: S,
    GatewayIP: G,
  }
  copy(msg.ClientHWAddr[0:], hwaddr)
  return
}


func (self Message)Marshal()(out []byte, err os.Error){
  out = make([]byte, DHCP_MAX_LEN)
  out[0] = self.Operation
  out[1] = self.HardwareType
  out[2] = byte(self.HardwareLen)
  out[3] = byte(self.Hops)
  binary.LittleEndian.PutUint32(out[4:8], self.Xid)
  binary.LittleEndian.PutUint16(out[8:10], self.Secs)
  binary.LittleEndian.PutUint16(out[10:12], uint16(self.Flags))
  copy(out[12:16], self.ClientIP.To4())
  copy(out[16:20], self.YourIP.To4())
  copy(out[20:24], self.ServerIP.To4())
  copy(out[24:28], self.GatewayIP.To4())
  copy(out[28:44], self.ClientHWAddr[0:])
  out[DHCP_MIN_LEN], out[DHCP_MIN_LEN+1],
  // DHCP "magic" signature
  out[DHCP_MIN_LEN+2], out[DHCP_MIN_LEN+3] = 0x63, 0x82, 0x53, 0x63

  pos := DHCP_MIN_LEN + 4
  for i := range(self.Options) {
    var bo []byte
    bo, err = MarshalOption(self.Options[i])
    if err != nil { break }
    copy(out[pos:], bo)
    pos += len(bo)
  }
  //copy(out[DHCP_MIN_LEN:], self.Options)

  return
}

const (
  BOOT_REQUEST byte = 0x01
  BOOT_REPLY   byte = 0x02
)

const (
  ETHERNET byte = 0x01
)

const DHCP_MAX_LEN = 576
const DHCP_MIN_LEN = 236
const (
  FLAG_BROADCAST uint16 = 0x80
)


func WriteMessage(c *net.UDPConn, msg *Message, dst net.IP)(err os.Error){
  out, err := msg.Marshal()
  if err != nil { return }
  _, err = c.WriteToUDP(out,  &net.UDPAddr{IP: dst, Port: 67})
  return
}

func ReadMessage(c *net.UDPConn)(msg *Message, src net.Addr, err os.Error){
  buff := make([]byte, DHCP_MAX_LEN)
  //log.Printf("Waiting for read")
  n, src, err := c.ReadFromUDP(buff)
  if err != nil { return }
  if n < DHCP_MIN_LEN {
    err = os.NewError("Invalid DHCP messge received (too small)")
    return
  }
  buff = buff[0:n]
  msg = &Message{
    Operation: buff[0],
    HardwareType: buff[1],
    HardwareLen: buff[2],
    Hops: buff[3],
    Xid: binary.LittleEndian.Uint32(buff[4:8]),
    Secs: binary.LittleEndian.Uint16(buff[8:10]),
    Flags: binary.LittleEndian.Uint16(buff[10:12]),
    ClientIP: net.IPv4(buff[12], buff[13], buff[14], buff[15]),
    YourIP: net.IPv4(buff[16], buff[17], buff[18], buff[19]),
    ServerIP: net.IPv4(buff[20], buff[21], buff[22], buff[23]),
    GatewayIP: net.IPv4(buff[24], buff[25], buff[26], buff[27]),
  }
  copy(msg.ClientHWAddr[0:16], buff[28:44])
  // We skip the magic bytes and assume for now.
  msg.Options, err = ParseOptions(buff[DHCP_MIN_LEN+4:])
  //log.Printf("Parsed %d options.", len(msg.Options))
  // TODO: Handle Option 52 extensions.
  return
}

func (self Message)OptionBytes(tag byte)(out []byte, err os.Error){
  for i := range(self.Options) {
    if self.Options[i].OptionType() == tag {
      out = self.Options[i].Bytes()
      break
    }
  }
  if out == nil { err = os.NewError("Option not found") }
  return
}

func (self Message)DHCPMessageType()(out byte, err os.Error){
  outb, err := self.OptionBytes(DHCPOPT_MESSAGE_TYPE)
  if err == nil {
    if len(outb) == 1 {
      out = outb[0]
    } else {
      err = os.NewError("Invalid DHCP message type received (wrong size)")
    }
  }
  return
}

func (self Message)ipOptionValue(tag byte)(ip net.IP, err os.Error){
  bits, err := self.OptionBytes(tag)
  if err == nil {
    if len(bits) ==4 || len(bits) == 16 {
      ip = net.IP(bits)
    } else {
      err = os.NewError("Don't know how to return a single IP at this size!")
    }
  }
  return
}

func (self Message)U32OptionValue(tag byte)(i uint32, err os.Error){
  var b []byte
  b, err = self.ipOptionValue(tag)
  if err == nil {
    if len(b) == 4 {
      i = binary.BigEndian.Uint32(b)
    } else {
      err = os.NewError("invalid length for u32 option")
    }
  }
  return
}


func (self Message)IP4OptionValue(tag byte)(ip net.IP, err os.Error){
  ip, err = self.ipOptionValue(tag)
  if err == nil {
    ip = ip.To4()
    if ip == nil { err = os.NewError("Address doesn't appear to be a valid IPv4 address.") }
  }
  return
}

func (self Message)IP4sOptionValue(tag byte)(ips []net.IP, err os.Error){
  bits, err := self.OptionBytes(tag)
  if len(bits) < 4 {
    err = os.NewError("Invalid length for net.IP[]")
  } else {
    if len(bits) % 4 == 0 {
       for i := 0; i*4+4 <= len(bits); i++ {
         ips = append(ips, net.IP(bits[i*4:i*4+4]))
       }
    } else {
      err = os.NewError("Invalid length for net.IP[]")
    }
  }
  return
}
