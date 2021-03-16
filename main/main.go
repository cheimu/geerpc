package main

import (
	"context"
	"geerpc"
	"geerpc/xclient"
	"log"
	"net"
	"sync"
	"time"
)

/* foo service and sum service */
type Foo int

type Args struct{ Num1, Num2 int }

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addrCh chan string) {
	var foo Foo
	// listen
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	// Now assume it is a rpc server
	server := geerpc.NewServer()
	// register service
	if err := server.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}
	// Now assume it is a http server
	// geerpc.HandleHTTP()
	// send the listerer's ip address into the channel
	addrCh <- l.Addr().String()
	server.Accept(l)
	// _ = http.Serve(l, nil)
}

func call(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// expect 2 - 5 timeout
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch1 := make(chan string)
	ch2 := make(chan string)
	// start two servers
	go startServer(ch1)
	go startServer(ch2)

	addr1 := <-ch1
	addr2 := <-ch2

	time.Sleep(time.Second)
	call(addr1, addr2)
	broadcast(addr1, addr2)
}

/*
	Main and Call before using load-balancer
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
