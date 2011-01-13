package main

import "dhcp"
import "flag"
import "log"
import "time"

func main(){
  ifname := flag.String("i","eth0","Interface name")
  flag.Parse()
  // current EC2 test host..
  ///c, err := dhcp.NewClientV4(*ifname, dhcp.ETHERNET, []byte{0x12,0x31,0x40,0x00,0x55,0x65})
  // default KVM host
  c, err := dhcp.NewClient(*ifname, dhcp.ETHERNET, []byte{0x54,0x52,0x00,0x12,0x34,0x56})
  if err != nil {
    log.Exitf("Couldn't create client: %v", err)
  }
  c.StartReader()
  log.Printf("Sending DHCPDISCOVER")
  err = c.Discover(nil)
  if err != nil {
    log.Exitf("Couldn't do discover: %v", err)
  }
  log.Printf("Collecting offers")
  time.Sleep(1000000000)
  //err = c.Request(*offer, []dhcp.Option{srvopt})
  //c.Release()
}
