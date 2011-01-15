package main

import "dhcp"
import "log"
import "flag"
import "fmt"
import "json"

func main(){
  ifname := flag.String("if","eth0","Interface to use")
  ifaddr := flag.String("mac","00:01:de:ad:be:ef","Mac to use ")
  flag.Parse()

  af, err := dhcp.NewDHCP4Finder(*ifname)
  if err != nil {
    log.Exitf("Cannot create finder for interface: %v", err)
  }
  mac := [6]byte{}
  _, err = fmt.Sscanf(*ifaddr, "%x:%x:%x:%x:%x:%x", &mac[0], &mac[1], &mac[2], &mac[3], &mac[4], &mac[5])
  if err != nil {
    log.Exitf("Cannot parse mac: %v", err)
  }
  offers := dhcp.AddressFinderResponse{}
  accepted := dhcp.Message{}
  err = af.DiscoverOffers(&dhcp.AddressFinderRequest{
    Timeout: 5000000000,
    MaximumOffers: 1,
    IfType: dhcp.ETHERNET,
    IfAddr: mac[0:6],
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
