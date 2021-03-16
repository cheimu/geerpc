package main

import (
	"context"
	"geerpc"
	"geerpc/app"
	"geerpc/registry"
	"geerpc/xclient"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func startRegistry(wg *sync.WaitGroup) {
	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal("network error:", err)
	}
	registry.HandleHTTP()
	wg.Done()
	if err := http.Serve(l, nil); err != nil {
		log.Fatal("http server error:", err)
	}
}

func startServer(registryAddr string, wg *sync.WaitGroup) {
	var foo app.Foo
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	server := geerpc.NewServer()
	if err := server.Register(&foo); err != nil {
		log.Fatal("service register error:", err)
	}
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	wg.Done()
	server.Accept(l)
}

func call(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			app.FooWrapper(xc, context.Background(), "call", "Foo.Sum", &app.Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			app.FooWrapper(xc, context.Background(), "broadcast", "Foo.Sum", &app.Args{Num1: i, Num2: i * i})
			// expect 2 - 5 timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			app.FooWrapper(xc, ctx, "broadcast", "Foo.Sleep", &app.Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	call(registryAddr)
	broadcast(registryAddr)
}

/*
	Main and Call before using Load-balancing and Registry
*/
// func call(addrCh chan string) {
// 	// dial and <-addr will get the ip address which already stored in channel when start server
// 	client, _ := geerpc.XDial("http@" + <-addrCh)
// 	defer func() { _ = client.Close() }()

// 	// wait to dail
// 	time.Sleep(time.Second)
// 	// send request & receive response
// 	var wg sync.WaitGroup
// 	for i := 0; i < 5; i++ {
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()
// 			args := &Args{Num1: i, Num2: i * i}
// 			var reply int
// 			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
// 				log.Fatal("call Foo.Sum error:", err)
// 			}
// 			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
// 		}(i)
// 	}
// 	wg.Wait()
// }

// func main() {
// 	log.SetFlags(0)
// 	ch := make(chan string)
// 	/* Why is this order? Why can't call startServer() first?
// 	Because we need keep running the process to make debug webpage online, and
// 	startServer(ch) will call handleHTTP() and will run the debug.ServeHTTP() forever,
// 	so need to use goroutine to run call(), and call startServer() in the end;

// 	If you Switch the order like following:
// 		startServer(ch)
// 		go call(ch)
// 	then call(ch) will never be called, because startServer(ch) will run all the time

// 	you can do following instead:
// 		go startServer(ch)
// 		go call(ch)
// 		time.Sleep(5 * time.Second)
// 	This will also work
// 	*/
// 	go call(ch)
// 	startServer(ch)
// }
