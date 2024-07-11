package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface {
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

func newsimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type loadBalancer struct {
	port            string
	roundrobincount int
	servers         []Server
}

func NewloadBalander(port string, servers []Server) *loadBalancer {
	return &loadBalancer{port: port,
		roundrobincount: 0,
		servers:         servers}
}

func handleErr(err error) {
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
func (s *simpleServer) Address() string { return s.addr }

func (s *simpleServer) IsAlive() bool { return true }

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request) {
	s.proxy.ServeHTTP(rw, req)
}
func (lb *loadBalancer) getNextAvailServer() Server {
	server := lb.servers[lb.roundrobincount%len(lb.servers)]

	for !server.IsAlive() {
		lb.roundrobincount++
		server = lb.servers[lb.roundrobincount%len(lb.servers)]
	}
	lb.roundrobincount++
	return server
}

func (lb *loadBalancer) serverProxy(rw http.ResponseWriter, req *http.Request) {
	target := lb.getNextAvailServer()
	fmt.Printf("forwarding request to address %q\n", target.Address())
	target.Serve(rw, req)

}

func main() {
	servers := []Server{
		newsimpleServer("https://www.facebook.com"),
		newsimpleServer("https://wwww.bing.com"),
		newsimpleServer("https://www.duckduckgo.com"),
	}
	lb := NewloadBalander("8000", servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request) {
		lb.serverProxy(rw, req)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("service is runnig at localhost:%s\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
