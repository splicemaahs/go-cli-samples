package main

import (
    "flag"
    "fmt"
    "net"
    "os"
    "time"

    "github.com/cakturk/go-netstat/netstat"
)

type Connection struct {
    Proto  string
    Source string
    Dest   string
    State  string
    Time   string
}

var (
    udp       = flag.Bool("udp", false, "display UDP sockets")
    tcp       = flag.Bool("tcp", false, "display TCP sockets")
    listening = flag.Bool("lis", false, "display only listening sockets")
    all       = flag.Bool("all", false, "display both listening and non-listening sockets")
    resolve   = flag.Bool("res", false, "lookup symbolic names for host addresses")
    ipv4      = flag.Bool("4", false, "display only IPv4 sockets")
    ipv6      = flag.Bool("6", false, "display only IPv6 sockets")
    help      = flag.Bool("help", false, "display this help screen")
)

const (
    protoIPv4 = 0x01
    protoIPv6 = 0x02
)

func main() {

    var activeConnections map[string]*Connection
    activeConnections = make(map[string]*Connection)
    flag.Parse()

    if *help {
        flag.Usage()
        os.Exit(0)
    }

    var proto uint
    if *ipv4 {
        proto |= protoIPv4
    }
    if *ipv6 {
        proto |= protoIPv6
    }
    if proto == 0x00 {
        proto = protoIPv4 | protoIPv6
    }

    if os.Geteuid() != 0 {
        fmt.Println("Not all processes could be identified, you would have to be root to see it all.")
    }

    for {
        //fmt.Printf("Proto %-23s %-23s %-12s %-16s\n", "Local Addr", "Foreign Addr", "State", "PID/Program name")

        if *udp {
            if proto&protoIPv4 == protoIPv4 {
                tabs, err := netstat.UDPSocks(netstat.NoopFilter)
                if err == nil {
                    displaySockInfo("udp", tabs, activeConnections)
                }
            }
            if proto&protoIPv6 == protoIPv6 {
                tabs, err := netstat.UDP6Socks(netstat.NoopFilter)
                if err == nil {
                    displaySockInfo("udp6", tabs, activeConnections)
                }
            }
        } else {
            *tcp = true
        }

        if *tcp {
            var fn netstat.AcceptFn

            switch {
            case *all:
                fn = func(*netstat.SockTabEntry) bool { return true }
            case *listening:
                fn = func(s *netstat.SockTabEntry) bool {
                    return s.State == netstat.Listen
                }
            default:
                fn = func(s *netstat.SockTabEntry) bool {
                    return s.State != netstat.Listen
                }
            }

            if proto&protoIPv4 == protoIPv4 {
                tabs, err := netstat.TCPSocks(fn)
                if err == nil {
                    displaySockInfo("tcp", tabs, activeConnections)
                }
            }
            if proto&protoIPv6 == protoIPv6 {
                tabs, err := netstat.TCP6Socks(fn)
                if err == nil {
                    displaySockInfo("tcp6", tabs, activeConnections)
                }
            }
        }
        time.Sleep(250 * time.Millisecond)
        // fmt.Printf("%#v", activeConnections)
        //for _, value := range activeConnections {
        //  // fmt.Println(fmt.Sprintf("Outside: %s - %s", value.Proto, value.Source))
        //  fmt.Printf("%-5s %-23.23s %-23.23s %-12s %-16s\n", value.Proto, value.Source, value.Dest, value.State, "")
        //}
    }
}

func displaySockInfo(proto string, s []netstat.SockTabEntry, connlist map[string]*Connection) {
    lookup := func(skaddr *netstat.SockAddr) string {
        const IPv4Strlen = 17
        addr := skaddr.IP.String()
        if *resolve {
            names, err := net.LookupAddr(addr)
            if err == nil && len(names) > 0 {
                addr = names[0]
            }
        }
        if len(addr) > IPv4Strlen {
            addr = addr[:IPv4Strlen]
        }
        return fmt.Sprintf("%s:%d", addr, skaddr.Port)
    }

    var modifiedlist map[string]*Connection
    modifiedlist = make(map[string]*Connection)
    var currentlist map[string]*Connection
    currentlist = make(map[string]*Connection)
    for _, e := range s {
        saddr := lookup(e.LocalAddr)
        daddr := lookup(e.RemoteAddr)
        mapkey := fmt.Sprintf("%s-%s-%s", proto, saddr, daddr)
        t := time.Now()
        currentlist[mapkey] = &Connection{Proto:proto, Source:saddr, Dest:daddr, State:fmt.Sprintf("%s", e.State), Time:fmt.Sprintf(t.Format("2006-01-02 15:04:05"))}
    }
    for _, e := range s {
        //p := ""
        //if e.Process != nil {
        //    p = e.Process.String()
        //}
        saddr := lookup(e.LocalAddr)
        daddr := lookup(e.RemoteAddr)
        // fmt.Printf("%-5s %-23.23s %-23.23s %-12s %-16s\n", proto, saddr, daddr, e.State, p)
        mapkey := fmt.Sprintf("%s-%s-%s", proto, saddr, daddr)
        // fmt.Printf("%s\n", mapkey)
        t := time.Now()
        if keyval, ok := connlist[mapkey]; !ok {
            connlist[mapkey] = &Connection{Proto:proto, Source:saddr, Dest:daddr, State:fmt.Sprintf("%s", e.State), Time:fmt.Sprintf(t.Format("2006-01-02 15:04:05"))}
            modifiedlist[mapkey] = &Connection{Proto:proto, Source:saddr, Dest:daddr, State:fmt.Sprintf("NEW -> %s", e.State), Time:fmt.Sprintf(t.Format("2006-01-02 15:04:05"))}
        } else {
           if keyval.State != fmt.Sprintf("%s",e.State) {
                modifiedlist[mapkey] = &Connection{Proto:proto, Source:saddr, Dest:daddr, State:fmt.Sprintf("%s -> %s", keyval.State, e.State), Time: fmt.Sprintf(t.Format("2006-01-02 15:04:05"))}
		connlist[mapkey].State = fmt.Sprintf("%s",e.State)
                connlist[mapkey].Time = fmt.Sprintf(t.Format("2006-01-02 15:04:05"))
	   } 
	}
        for key, conn := range connlist {
		if _, ok := currentlist[key]; !ok {
                    modifiedlist[key] = &Connection{Proto:conn.Proto, Source:conn.Source, Dest:conn.Dest, State:fmt.Sprintf("%s -> CLOSED", conn.State), Time:fmt.Sprintf(t.Format("2006-01-02 15:04:05"))}
                    delete(connlist, key)
		}
	}
    }
    //for _, value := range connlist {
    //    fmt.Println(fmt.Sprintf("%s - %s", value.Proto, value.Source))
    //}
    for _, value := range modifiedlist {
          fmt.Printf("%-16s %-5s %-23.23s %-23.23s %-12s\n", value.Time, value.Proto, value.Source, value.Dest, value.State)
    }
}

