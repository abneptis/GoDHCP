package dhcp

import "log"
import "os"
import "strconv"

type OfferAcceptor interface {
  AcceptOffer(Message)(os.Error)
}

type nilAcceptor struct {}
type defaultAcceptor struct {}

func (self nilAcceptor)AcceptOffer(msg Message)(err os.Error){
  t, err := msg.DHCPMessageType()
  if err == nil {
    if t != DHCPACK {
      err = os.NewError("nilAcceptor: message is NOT an DHCP ACK !" + strconv.Itoa(int(t)))
    } else {
      log.Printf("nilAcceptor: Accepted message")
    }
  }
  return
}

func (self defaultAcceptor)AcceptOffer(msg Message)(err os.Error){
  ackmap := map[string]interface{}{
    "YourIP" : msg.YourIP,
    "ServerIP" : msg.ServerIP,
    "GatewayIP" : msg.GatewayIP,
    "ClientIP" : msg.ClientIP,
  }
  optmap := map[string]interface{}{}
  for oi := range(msg.Options){
    optType := msg.Options[oi].OptionType()
    str := OptionTypeStrings[optType]
    if str == "" {
      log.Printf("Warning: Unknown option type: %d", optType)
      continue
    }
    switch optType {
      case OPT_PAD, OPT_END: // Do nothing with these in terms of marshalling.
      case OPT_DOMAIN_NAME:
        optmap[str] = string(msg.Options[oi].Bytes())
      case OPT_DEFAULT_GATEWAY, OPT_SUBNET_MASK, DHCPOPT_SERVER_ID:
        optmap[str], err = msg.IP4OptionValue(optType)
        if err != nil { break }
      case DHCPOPT_LEASE_TIME, DHCPOPT_T1, DHCPOPT_T2:
        optmap[str], err = msg.U32OptionValue(optType)
        if err != nil { break }
      case OPT_DOMAIN_NAME_SERVERS:
        optmap[str], err = msg.IP4sOptionValue(optType)
      default:
        optmap[str] = msg.Options[oi].Bytes()
    }
    if err != nil {
      log.Printf("Warning: Upper marshalling error: %v", err)
    }
  }
  if err == nil { ackmap["Options"] = optmap }
  log.Printf("Debug: Received ACK : %+v [%+v]", ackmap, ackmap["Options"])
  return

}
