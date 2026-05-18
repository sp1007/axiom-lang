# p16-t06: std.net — Network I/O

## Purpose
Implement TCP/UDP networking for AXIOM programs, integrating with the async runtime for non-blocking connections, supporting both client and server use cases.

## Context
`std.net` provides the low-level networking primitives. It wraps POSIX sockets (Linux/macOS) and Winsock (Windows) with AXIOM's ownership model and async integration. TCP is the primary target; UDP is secondary.

## Inputs
- I/O event loop from p15-t08
- OS socket APIs: socket/connect/accept/send/recv (POSIX), WSASocket (Windows)
- AXIOM async/await from p15-t06

## Outputs
- `stdlib/net/tcp.ax` — TcpStream, TcpListener
- `stdlib/net/udp.ax` — UdpSocket
- `stdlib/net/addr.ax` — IpAddr, SocketAddr parsing

## Dependencies
- p15-t08: io-event-loop — non-blocking socket registration
- p16-t04: std-io — IOError type reuse

## Detailed Requirements

```axiom
# stdlib/net/addr.ax
type IpAddr:
    V4(u8, u8, u8, u8)
    V6(u8, u8, u8, u8, u8, u8, u8, u8, u8, u8, u8, u8, u8, u8, u8, u8)

type SocketAddr:
    var ip:   IpAddr
    var port: u16

    fn parse(s: str) -> Result[SocketAddr, str]  # "127.0.0.1:8080"
    fn to_str(self) -> str

# stdlib/net/tcp.ax
type TcpStream:
    var fd: i32

    async fn connect(addr: SocketAddr) -> Result[TcpStream, IOError]
    async fn read(mut self, buf: []u8) -> Result[u32, IOError]
    async fn write(mut self, buf: []u8) -> Result[u32, IOError]
    async fn write_all(mut self, buf: []u8) -> Result[void, IOError]
    fn set_nodelay(mut self, enabled: bool) -> Result[void, IOError]
    fn close(mut self)

type TcpListener:
    var fd: i32

    fn bind(addr: SocketAddr) -> Result[TcpListener, IOError]
    fn listen(mut self, backlog: i32) -> Result[void, IOError]
    async fn accept(mut self) -> Result[TcpStream, IOError]
    fn close(mut self)

# stdlib/net/udp.ax
type UdpSocket:
    var fd: i32

    fn bind(addr: SocketAddr) -> Result[UdpSocket, IOError]
    async fn send_to(mut self, buf: []u8, addr: SocketAddr) -> Result[u32, IOError]
    async fn recv_from(mut self, buf: []u8) -> Result[(u32, SocketAddr), IOError]
    fn close(mut self)
```

Non-blocking sockets:
- Set `O_NONBLOCK` (POSIX) or `FIONBIO` (Windows) on all sockets.
- Register with I/O event loop via `ax_ioloop_register(fd, EPOLLIN|EPOLLOUT)`.
- async calls suspend via `AxFuture`; event loop resolves on I/O readiness.

## Implementation Steps

1. Create `stdlib/net/addr.ax` — IpAddr, SocketAddr parsing.
2. Create `stdlib/net/tcp.ax` — TcpStream and TcpListener.
3. Implement socket creation as non-blocking with O_NONBLOCK.
4. Wire async read/write to I/O event loop futures.
5. Create `stdlib/net/udp.ax`.
6. Write integration tests using localhost connections.

## Test Plan
- `TestTcpConnectAccept`: client connects → server accepts → exchange "hello"
- `TestTcpReadWrite`: send 1MB → receive 1MB → byte-identical
- `TestTcpListener1000`: 1000 concurrent connections to listener
- `TestSocketAddrParse`: "127.0.0.1:8080" parses correctly
- `TestUdpSendRecv`: UDP packet round-trip on localhost

## Validation Checklist
- [ ] All sockets set to non-blocking mode
- [ ] TcpStream/UdpSocket auto-closed by CTGC
- [ ] async connect handles EINPROGRESS correctly
- [ ] IPv6 addresses parsed and handled

## Acceptance Criteria
- Simple HTTP server handling 10K concurrent connections with 4 worker threads

## Definition of Done
- [ ] `stdlib/net/tcp.ax` and `stdlib/net/udp.ax` implemented
- [ ] localhost TCP tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| EINPROGRESS on async connect (non-blocking) | Wait for EPOLLOUT, then check SO_ERROR with getsockopt |
| Windows Winsock differences | Abstract behind platform-specific C shims |

## Future Follow-up Tasks
- TLS wrapper using OpenSSL or rustls-ffi
- HTTP client/server (stdlib/http)
