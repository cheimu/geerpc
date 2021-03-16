# Geerpc
Implementation of a simple remote procedure call <br>
Reference: https://github.com/geektutu/7days-golang/tree/master/gee-rpc

## Day 1: Runnable RPC Framework<br> ##
Initialization<br>
`codec.init()` and `NewServer()` <br>
Client side
1. make(chan string)
2. `startServer()`<br>
   |<br>
   |--- `net.Listen()`<br>
   |--- server side `server.Accept()`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `DefaultServer.Accept()`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `server.ServeConn()`(STAY HERE) <br> 
3. `net.Dial()`<br>
4. sleep (wait until connection built by `net.Dail()`)
5. send Option
6. `NewGobCodec()`
7. Prepare Header
8. Write<br>
   |<br>
   |--- server side Receive Option (CONTINUE the `server.ServeConn()`)<br>
   |--- server side check magic number in Option<br>
   |--- server side `NewGobCodec()` which is decided in Option <br>
   |--- server side `server.serveCodec()`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `server.readRequest()`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `server.readRequestHeader())`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `GobCodec.ReadHeader()`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `GobCodec.ReadBody()`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `server.handleRequest()`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `server.SendResponse())`<br>
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|--- server side `GobCodec.Write()`<br>
9. `GobCodec.ReadHeader()`
10. `GobCodec.ReadBody()`
11. `conn.Close()`

## Day 2: Concurrent Client Side <br> ## 
`Main()`<br>
1. Create client instance and Receive: <br>
`client = geerpc.Dail()` -> `client.NewClient()` -> send option and `client.newClientCodec()` -> `go client.receive()` and return client -> `client.cc.ReadHeader()` then `client.removeCall()`, then `client.cc.ReadBody()`, and finally `client.terminateCalls()`<br>
2. Send: <br>
WaitGroup's things, `Add()`, `Done()` and `Wait()`, and `go client.Call()` -> `client.Go()` -> `client.send()` -> `client.registerCall()` and `client.cc.Write(&client.header, call.Args)`

## Day 3: Service Call using Reflect <br> ## 
Places that need modification<br>
1. `startServer()` -> `geerpc.Register(&service_name)` -> `server.Register()` -> `newServer()` -> `service.registerMethods()` -> check `isExportedOrBuiltinType()`
2. `server.ReadRequest()` -> `req.svc, req.mtype, err = server.findService(h.ServiceMethod)`, `req.argv = req.mtype.newArgv()`, `req.replyv = req.mtype.newReplyv()`, `argvi := req.argv.Interface()` and finally `cc.ReadBody(argvi)`
3. `server.handleRequest()` -> `req.svc.call(req.mtype, req.argv, req.replyv)` -> `service.call()`

## Day 4: Handle Timeouts <br> ##
__The technique here is setup a signal and its channel, and use select to check. if `case <-timeout_channel` happened first means timeouts; if signal happened first means no timeouts.__<br>

_For Client side timeouts_: <br>
1. Setup connection to server timeouts: <br>
`Dial()` -> `client.dailTimeout()` -> `conn, err := net.DialTimeout()` and `client, err := f(conn, opt), ch <- clientResult{client: client, err: err}`. Then do select technique described above using `time.After(opt.ConnectTimeout)`
2. Send request timeouts &
3. Wait for response too long got timeoutsand &
4. Receive response got timeouts : <br>
In `Call()` use `ctx context.Context`. Do select technique described above using `ctx-Done()` 

_For Server side timeouts_:<br>
1. when receive request &
2. when call services &
3. when send request:<br>
In `server.handleRequest()` uses two signals, `called` and `sent`. `called` indicates that started a goroutine to call and `req.svc.call(req.mtype, req.argv, req.replyv)` is executed but `server.sendResponse()` isn't, while `sent` incidates `server.sendResponse()` is executed.<br>
Then do select technique described above using `time.After(timeout)`

## Day 5: Handle Timeouts <br> ##
The idea is Client side treats server as HTTP server and send `CONNECT` request, then Server side will return a HTTP response to indicate the success of connection. Then each message in HTTP will be hijacked by GeeRPC and executed.<br>
Then when dial, call `XDial()` which will choose to `DialHTTP()` or `Dial()`

`main()`<br>
        &nbsp;&nbsp;|<br>
        &nbsp;&nbsp;|->`startServer()` -> `handleHTTP()` -> `DefaultServer.handleHTTP()`->`http.Handle(defaultRPCPath, server)` and `http.Handle(defaultDebugPath, debugHTTP{server})` -> `server.ServeHTTP()` and `debugHTTP.ServeHTTP()` <br>
        &nbsp;&nbsp;|-> `go call()` -> `XDial("http@" + <-addrCh)`

## Day 6: Load-Balancing <br> ##
`main()`<br>
        &nbsp;&nbsp;|<br>
        &nbsp;&nbsp;|->`d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})`<br>
        &nbsp;&nbsp;|->`xc := xclient.NewXClient(d, xclient.<load_balancing_mode>, nil)`<br>
        
__For single call__: <br>
`xc.Call()` -> `rpcAddr, err := sc.d.Get(<load_balancing_mode>)` then `xc.call(rpcAddr)` -> `xc.dial(rpcAddr, ctx, serviceMethod, args, Reply)` -> `client = XDial(rpcAddr, xc.opt)` then `client.Call()` <br>
__For broadcast__: <br>
`xc.Broadcast()` -> `servers, err := xc.d.GetAll()`, then for each server in servers: if `reply` is not nil: `clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()`, `xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply)`, and `reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())`

