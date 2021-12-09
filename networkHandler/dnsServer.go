package networkHandler

import (
	"context"
	"log"
	"math/rand"
	"net"
	"regexp"
	"sync"
	"time"

	"github.com/miekg/dns"
)

var dataCH = make(map[string]string)
var datamx = &sync.Mutex{}
var flush = make(chan struct{})

type DNSHandler struct{}

func (*DNSHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		address, ok := haveIT(domain)
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address),
			})
		}
	}
	w.WriteMsg(&msg)
}

func flushCH() {
	for {
		<-flush
		log.Printf("dropping memory cache .... %v recored", len(dataCH))
		dataCH = make(map[string]string)
	}
}

func haveIT(domain string) (string, bool) {
	for k, v := range dataCH {
		if ok, _ := regexp.MatchString(k, domain); ok {
			return v, true
		}
	}

	if addr, ok := askUpstr(domain); ok {
		datamx.Lock()
		dataCH[domain] = addr
		datamx.Unlock()
		return addr, true
	}
	return "err", false
}

func timeCh() {
	for {
		time.Sleep(time.Second*time.Duration(rand.Intn(120)) + 90)
		flush <- struct{}{}
	}
}

func memCH() {
	for {
		time.Sleep(time.Millisecond * 100)
		if len(dataCH) > 900000 {
			flush <- struct{}{}
		}

	}
}

func RunDNS(ctx context.Context, addr string) {
	go flushCH()
	go timeCh()
	go memCH()
	go func() {
		srv := &dns.Server{Addr: addr, Net: "udp"}
		srv.Handler = &DNSHandler{}
		dns.NewServeMux()
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set udp DNS listener %s\n", err.Error())
		}
	}()

	go func() {
		srv := &dns.Server{Addr: addr, Net: "tcp"}
		srv.Handler = &DNSHandler{}
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Failed to set tcp DNS listener %s\n", err.Error())
		}
	}()
	<-ctx.Done()
}

func askUpstr(domain string) (string, bool) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		},
	}
	ip, err := r.LookupHost(context.Background(), domain)
	if err != nil {
		log.Println(err)
		return "err", false
	}
	if len(ip) != 0 {
		return ip[0], true
	}
	return "err", false
}
