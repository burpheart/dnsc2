# dnsc2

一个通过dns双向通信简单的c2

# 功能支持

DoT  (Dns Over TLS) TLS加密传输dns流量

客户端返回数据分包 重组包

客户端多线程发包 减少回传数据时间

学习go 练手写的一个小工具   勉强能用的状态)


# HTTP服务

HTTP服务默认监听在 0.0.0.0:2333

/dns/Command 添加一个任务到队列 ?cmd=任务数据&aimid=目标客户端id(空为不指定客户端)&type=任务类型 (1 cmd 2 sleep 3 exit  默认 1)

/dns/GetTasks 获取任务列表

/dns/GetResults 获取结果列表

/dns/GetClients 获取客户端列表

## 仅用于DNS隧道的技术研究
