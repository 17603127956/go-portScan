//go:build !nosyn

package syn

import (
	"errors"
	"fmt"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"github.com/jackpal/gateway"
	"github.com/libp2p/go-netroute"
	"net"
)

// GetDevByIp get dev name by dev ip (use pcap)
func GetDevByIp(ip net.IP) (devName string, err error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return
	}
	for _, d := range devices {
		for _, address := range d.Addresses {
			_ip := address.IP.To4()
			if _ip != nil && _ip.IsGlobalUnicast() && _ip.Equal(ip) {
				return d.Name, nil
			}
		}
	}
	return "", errors.New("can not find dev")
}

// GetIfaceMac get interface mac addr by interface ip (use golang net)
func GetIfaceMac(ifaceAddr net.IP) (src net.IP, mac net.HardwareAddr) {
	interfaces, _ := net.Interfaces()
	for _, iface := range interfaces {
		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				if addr.(*net.IPNet).Contains(ifaceAddr) {
					return addr.(*net.IPNet).IP, iface.HardwareAddr
				}
			}
		}
	}
	return nil, nil
}

// GetMacByGw get srcIp srcMac devname by gw
func GetMacByGw(gw net.IP) (srcIp net.IP, srcMac net.HardwareAddr, devname string, err error) {
	srcIp, srcMac = GetIfaceMac(gw)
	if srcIp == nil {
		err = errors.New("can not find this dev by gw")
		return
	}
	srcIp = srcIp.To4()
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return
	}
	for _, d := range devices {
		if len(d.Addresses) > 0 && d.Addresses[0].IP.String() == srcIp.String() {
			devname = d.Name
			return
		}
	}
	err = errors.New("can not find this dev")
	return
}

// GetRouterV4 get ipv6 router by dst ip
func GetRouterV4(dst net.IP) (srcIp net.IP, srcMac net.HardwareAddr, gw net.IP, devName string, err error) {
	// 同网段
	srcIp, srcMac = GetIfaceMac(dst)
	if srcIp == nil {
		var r routing.Router
		r, err = netroute.New()
		if err == nil {
			var iface *net.Interface
			iface, gw, srcIp, err = r.Route(dst)
			if err == nil {
				if iface != nil {
					srcMac = iface.HardwareAddr
				} else {
					_, srcMac = GetIfaceMac(srcIp)
				}
			}
		}
		if err != nil {
			// 取第一个默认路由
			gw, err = gateway.DiscoverGateway()
			if err == nil {
				srcIp, srcMac = GetIfaceMac(gw)
			}
		}
	}
	gw = gw.To4()
	srcIp = srcIp.To4()
	devName, _ = GetDevByIp(srcIp)
	if srcIp == nil || err != nil || srcMac == nil {
		if err == nil {
			err = fmt.Errorf("err")
		}
		return nil, nil, nil, "", fmt.Errorf("no router, %s", err)
	}
	return
}
