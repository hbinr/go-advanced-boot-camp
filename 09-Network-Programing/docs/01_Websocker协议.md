
## 概念
> WebSocket 是一种网络传输协议，可在单个 TCP 连接上进行全双工通信，位于 OSI 模型的应用层。 WebSocket 协议在 2011 年由 IETF 标准化为 RFC 6455，后由 RFC 7936 补充规范。Web IDL 中的 WebSocket API 由 W3C 标准化。

> WebSocket 使得客户端和服务器之间的数据交换变得更加简单，允许服务端主动向客户端推送数据。在 WebSocket API 中，**浏览器和服务器只需要完成一次握手**，两者之间就可以创建**持久性的连接**，并进行**双向**数据传输。

WebSocket 是一种与 HTTP 不同的协议。两者都位于 OSI 模型的应用层，并且都依赖于传输层的 TCP 协议。虽然它们不同，但 RFC 6455 规定：“WebSocket 设计为通过 80 和 443 端口工作，以及支持 HTTP 代理和中介”，从而使其与 HTTP 协议兼容。 为了实现兼容性，WebSocket 握手使用 HTTP Upgrade 头从 HTTP 协议更改为 WebSocket 协议。

WebSocket 协议支持 Web 浏览器（或其他客户端应用程序）与 Web 服务器之间的交互，具有较低的开销，便于实现客户端与服务器的实时数据传输。 服务器可以通过标准化的方式来实现，而无需客户端首先请求内容，并允许消息在保持连接打开的同时来回传递。通过这种方式，可以在客户端和服务器之间进行双向持续交互。 通信默认通过 TCP 端口 80 或 443 完成。

大多数浏览器都支持该协议，包括 Google Chrome、Firefox、Safari、Microsoft Edge、Internet Explorer 和 Opera。

与 HTTP 不同，WebSocket 提供全双工通信。此外，WebSocket 还可以在 TCP 之上启用消息流。TCP 单独处理字节流，没有固有的消息概念。 在 WebSocket 之前，使用 Comet 可以实现全双工通信。但是 Comet 存在 TCP 握手和 HTTP 头的开销，因此对于小消息来说效率很低。WebSocket 协议旨在解决这些问题。

WebSocket 协议规范将 ws（WebSocket）和 wss（WebSocket Secure）定义为两个新的统一资源标识符（URI）方案，分别对应明文和加密连接。除了方案名称和片段 ID（不支持#）之外，其余的 URI 组件都被定义为此 URI 的通用语法。

## WebSocket 的优点

- 了解了 WebSocket 是什么，那 WebSocket 有哪些优点？这里总结如下：
- 较少的控制开销。在连接创建后，服务器和客户端之间交换数据时，用于协议控制的数据包头部相对较小。在不包含扩展的情况下，对于服务器到客户端的内容，此头部大小只有 2 至 10 字节（和数据包长度有关）；对于客户端到服务器的内容，此头部还需要加上额外的 4 字节的掩码。相对于 HTTP 请求每次都要携带完整的头部，此项开销显著减少了。
- 更强的实时性。由于协议是全双工的，所以服务器可以随时主动给客户端下发数据。相对于 HTTP 请求需要等待客户端发起请求服务端才能响应，延迟明显更少；即使是和 Comet 等类似的长轮询比较，其也能在短时间内更多次地传递数据。
- 保持连接状态。Websocket 是一种有状态的协议，通信前需要先创建连接，之后的通信就可以省略部分状态信息了。而 HTTP 请求可能需要在每个请求都携带状态信息（如身份认证等）。
- 更好的二进制支持。Websocket 定义了二进制帧，相对 HTTP，可以更轻松地处理二进制内容。
- 可以支持扩展。Websocket 定义了扩展，用户可以扩展协议、实现部分自定义的子协议。如部分浏览器支持压缩等。
- 更好的压缩效果。相对于 HTTP 压缩，Websocket 在适当的扩展支持下，可以沿用之前内容的上下文，在传递类似的数据时，可以显著地提高压缩率。
