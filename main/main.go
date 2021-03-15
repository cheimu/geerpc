package main

import (
	"context"
	"geerpc"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

/* foo service and sum service */
type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addrCh chan string) {
	var foo Foo
	// register service
	if err := geerpc.Register(&foo); err != nil {
		log.Fatal("register error:", err)
	}
	// listen
	l, err := net.Listen("tcp", ":9999")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	// Now assume it is a http server
	geerpc.HandleHTTP()
	// send the listerer's ip address into the channel
	addrCh <- l.Addr().String()
	// geerpc.Accept(l)
	_ = http.Serve(l, nil)
}

func call(addrCh chan string) {
	// dial and <-addr will get the ip address which already stored in channel when start server
	client, _ := geerpc.XDial("http@" + <-addrCh)
	defer func() { _ = client.Close() }()

	// wait to dail
	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch := make(chan string)
	/* Why is this order? Why can't call startServer() first?
	Because we need keep running the process to make debug webpage online, and
	startServer(ch) will call handleHTTP() and will run the debug.ServeHTTP() forever,
	so need to use goroutine to run call(), and call startServer() in the end;

	If you Switch the order like following:
		startServer(ch)
		go call(ch)
	then call(ch) will never be called, because startServer(ch) will run all the time

	you can do following instead:
		go startServer(ch)
		go call(ch)
		time.Sleep(5 * time.Second)
	This will also work
	*/
	go call(ch)
	startServer(ch)
}
