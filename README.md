场景
多个节点 ，举例： 比如有3个节点， ABC
ABC之间的网络。可能通。也可能只和另一个通

现在要在ABC上跑一个程序ffmesh，  提供http，tcp，udp等平等网络服务

ABC上的ffmesh运行  业务程序注册上来, 比如A上面收到本地auth微服务的注册
{
    "host":"auth",
    "port":"8080"
    "type":"http",
    "path":[
        "/login",
        "/logout"
    ]
}

此时abc上的程序， 访问http://auth/login， 能请求到A节点上面的auth微服务的login接口
mesh网络外层人员通过浏览器/或者api调用  访问ABC上的暴露端口80/443  
如用curl访问  http://A(或B,C)的ip/auth/login  也能访问到A节点上面的微服务auth的login接口



例子中是个小规模的  网络mesh，  实际可能会更多

每个节点上的业务服务  会注册到ffmesh上
ffmesh之间 会互联互通（ffmesh，只手动连上其他一个）

达到效果（模拟举例）：
1.B上的服务  可以通过调用本地ffmesh的端口访问到C上的mysql
2.公网用户访问A的公网IP能访问到C上的web页面
3.C上的局域网通过浏览器打开C的80端口能访问到A上面的web页面
4.如果大家都注册到A上  如果BC互相发送请求，他们能通就自己通，无法通就通过A代理
5.ffmesh启动时，可以填写另一个ffmesh的地址，主动过加入mesh网络


技术： libp2p
开发语言: golang