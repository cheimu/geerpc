# rpc
Implementation of a simple remote procedure call <br>
Reference: https://github.com/geektutu/7days-golang/tree/master/gee-rpc

## Day 1:<br> ##
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
3. `net.Dial`<br>
4. sleep (wait for dailing finished)
5. send Option
6. NewGobCodec()
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
9. GobCodec.ReadHeader()
10. GobCodec.ReadBody()
11. `conn.Close()`
