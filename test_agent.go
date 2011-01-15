package main

import "dhcp"
import "log"
import "json"

func main(){
  af, err := dhcp.NewDHCP4Finder("br0")
  if err != nil {
    log.Exitf("Cannot create finder for interface: %v", err)
  }
  offers := dhcp.AddressFinderResponse{}
  accepted := dhcp.Message{}
  err = af.DiscoverOffers(&dhcp.AddressFinderRequest{
    Timeout: 5000000000,
    MaximumOffers: 1,
    IfType: dhcp.ETHERNET,
    IfAddr: []byte{0,0x90,0xf5,0x9c,0x8e,0xf1},
  }, &offers)
  if err != nil {
    log.Exitf("Problem receiving offers: %v", err)
  }
  if len(offers) == 0 {
    log.Exitf("No DHCP Offers received")
  }
  ob, err := json.Marshal(offers)
  if err != nil {
    log.Exitf("Unable to marshal response")
  }
  err = af.AcceptOffer(&dhcp.AddressAcceptRequest{
   Offer: offers[0],
   Timeout: 5000000000,
  },&accepted)
  if err != nil {
    log.Exitf("No ACK received")
  }
  ob, err = json.Marshal(accepted)
  if err != nil {
    log.Exitf("Unable to marshal response")
  }
  log.Printf("%s", ob)
}
