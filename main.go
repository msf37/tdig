package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/miekg/dns"
)

var (
	domain    = flag.String("domain", "", "The domain to query")
	server    = flag.String("server", "", "The DNS server to use")
	qType     = flag.String("type", "A", "The query type (A, MX, NS, etc.)")
	recursion = flag.Bool("recursion", true, "Enable or disable recursion")
	suite     = flag.String("suite", "", "Specify tls 1.3 cipher suite")
)

var suites map[string]uint16

// TLS 1.3 cipher suites.
func initSuite() {
	suites = make(map[string]uint16)

	suites["TLS_AES_128_GCM_SHA256"] = 0x1301
	suites["TLS_AES_256_GCM_SHA384"] = 0x1302
	suites["TLS_CHACHA20_POLY1305_SHA256"] = 0x1303
	suites["TLS_AES_128_CCM_8_SHA256"] = 0x1305
	suites["TLS_AES_128_CCM_SHA256"] = 0x1304
}

func main() {
	flag.Parse()

	initSuite()

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

	var config *tls.Config

	config = &tls.Config{
		ServerName:         *server,
		InsecureSkipVerify: true,
	}

	if *suite != "" {
		s, ok := suites[*suite]
		if !ok {
			fmt.Println("Invalid cipher suites. please choose one of the list:")
			fmt.Println("	- TLS_AES_128_GCM_SHA256")
			fmt.Println("	- TLS_AES_256_GCM_SHA384")
			fmt.Println("	- TLS_CHACHA20_POLY1305_SHA256")
			fmt.Println("	- TLS_AES_128_CCM_8_SHA256")
			fmt.Println("	- TLS_AES_128_CCM_SHA256")
			os.Exit(0)
		}

		config = &tls.Config{
			ServerName:               *server,
			InsecureSkipVerify:       true,
			PreferServerCipherSuites: true,
			CipherSuites:             []uint16{s},
			MinVersion:               tls.VersionTLS13,
			MaxVersion:               tls.VersionTLS13,
		}
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
