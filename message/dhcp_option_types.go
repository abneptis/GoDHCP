package dhcp

const (
  DHCPOPT_REQUEST_IP byte = iota + 50 // 0x32, 4, net.IP
  DHCPOPT_LEASE_TIME // 0x33, 4, uint32
  DHCPOPT_EXT_OPTS   // 0x34, 1, 1/2/3
  DHCPOPT_MESSAGE_TYPE // 0x35, 1, 1-7
  DHCPOPT_SERVER_ID // 0x36, 4, net.IP
  DHCPOPT_PARAMS_REQUEST // 0x37, n, []byte
  DHCPOPT_MESSAGE // 0x38, n, string
  DHCPOPT_MAX_DHCP_SIZE // 0x39, 2, uint16
  DHCPOPT_T1 // 0x3a, 4, uint32
  DHCPOPT_T2 // 0x3b, 4, uint32
  DHCPOPT_CLASS_ID // 0x3c, n, []byte
  DHCPOPT_CLIENT_ID // 0x3d, n >=  2, []byte
)

