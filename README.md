请参考原版文档以及说明，与原版不兼容。主要修正包括：

1. 移除了一些不需要的模块，如mongo的支持等，这些直接用第三方库即可；
2. 将自己实现的log模块改为zap的，性能更好并支持json格式的日志；
3. 将websocket的RemoteAddr返回值改为透过代理的（如果存在）；
4. 加上go mod支持，修改版本号为规范格式；
5. protobuf消息支持自定义ID；
6. 增加了大量utils函数；
7. 增加了udp支持；将TCP的协议解析器和路由抽象出来，允许自行定义；
8. 允许动态加载各module，允许热重启所有module；
9. 增加了tcp和websocket的通用客户端；
10. 关闭信号由`chan bool`改为`chan struct{}`；
