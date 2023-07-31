package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

// server interface
type Server interface {
	Address() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request) 
}

// the struct defining the structure of a server
type baseServer struct {
	address 	string
	proxy 		*httputil.ReverseProxy
}

// function defining the working of a server
func theServer(address string) *baseServer {
	serverUrl, err := url.Parse(address)
	handleError(err) 

	return &baseServer{
		address: address,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

// the strcut defining the structure of a load balancer
type LoadBalancer struct {
	port 				string
	roundRobinCount 	int
	servers 			[]Server 
}

// function creating & defining the working of a load balancer
func loadBalancer(port string, servers []Server) *LoadBalancer {

	return &LoadBalancer{
		port: port,
		roundRobinCount: 0,
		servers: servers,
	}
}

// the central error handling function
func handleError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// method to return the address 
func (serv *baseServer) Address() string {
	return serv.address
}

// method to check server is alive or not 
func (serv *baseServer) IsAlive() bool {
	return true
}

// method to check server is alive or not & serving through this reverse proxy
func (serv *baseServer) Serve(w http.ResponseWriter, r *http.Request) {
	serv.proxy.ServeHTTP(w, r)
}

// go between servers and check which one is alive to pass the request
func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roundRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

// serveProxy method to get the next available server and choose it as a target to pass the request
func (lb *LoadBalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("Forwarding request to address %q\n", targetServer.Address())
	targetServer.Serve(w, r)
}


func main() {

	// a slice of servers
	servers := []Server{
		theServer("https://www.google.com"),
		theServer("https://www.bing.com"),
		theServer("https://www.duckduckgo.com"),
		theServer("https://www.search.brave.com"),
	}

	// creating load balancer and sending all the above servers
	lb := loadBalancer("8000", servers) // loadBalancer(port string, servers []Server)
	handleRedirect := func (w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w, r)
	}
	http.HandleFunc("/", handleRedirect)

	fmt.Printf("Server is serving request at port: %s\n", lb.port)
	http.ListenAndServe(":" + lb.port, nil)
}