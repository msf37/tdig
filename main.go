package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

var (
	domain    = flag.String("domain", "", "The domain to query")
	server    = flag.String("server", "", "The DNS server to use")
	qType     = flag.String("type", "A", "The query type (A, MX, NS, etc.)")
	recursion = flag.Bool("recursion", true, "Enable or disable recursion")
)

func main() {
	flag.Parse()

	if *domain == "" {
		fmt.Println("Please provide a domain to query.")
		return
	}

	if *server == "" {
		// Use the default DNS resolver configured on the system
		addrs, err := net.LookupHost("8.8.8.8")
		if err != nil || len(addrs) == 0 {
			fmt.Println("Failed to find a DNS server.")
			return
		}
		*server = addrs[0]
	}

	config := &tls.Config{
		ServerName:         *server,
		InsecureSkipVerify: true,
	}
	dialer := &net.Dialer{
		Timeout: time.Second * 10,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(*server, "853"), config)
	if err != nil {
		fmt.Printf("Error connecting to DNS server: %v\n", err)
		return
	}
	defer conn.Close()

	c := new(dns.Client)
	c.Net = "tcp-tls"

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(*domain), dns.StringToType[*qType])
	m.RecursionDesired = *recursion

	r, _, err := c.ExchangeWithConn(m, &dns.Conn{Conn: conn})
	if err != nil {
		fmt.Printf("Error querying DNS server: %v\n", err)
		return
	}

	if len(r.Answer) == 0 {
		fmt.Println("No results found.")
		return
	}

	for _, ans := range r.Answer {
		fmt.Printf("%v\n", ans)
	}
}
