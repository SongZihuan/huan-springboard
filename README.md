# Huan-Springboard
## 介绍
简单的 TCP 端口转发服务，使用了 PROXY 转发协议（使用了Version 1）。

## 如何配置
### 命令行参数
```text
Usage of HSBv1.exe:
  --help
  --h
          Show usage of HSBv1.exe. If this option is set, the backend service
          will not run.

  --version
  --v
          Show version of HSBv1.exe. If this option is set, the backend service
          will not run.

  --license
  --l
          Show license of HSBv1.exe. If this option is set, the backend service
          will not run.

  --report
  --r
          Show how to report questions/errors of HSBv1.exe. If this option is
          set, the backend service will not run.

  --config string
  --c string
          The location of the running configuration file of the backend service.
          The option is a string, the default value is config.yaml in the
          running directory.

  --output-config string
          The location of the reverse output after the backend service running
          configuration file is parsed. The option is a string and the default
          is config.output.yaml in the running directory.

  --not-auto-reload
          Disable automatic detection of configuration file changes and
          reloading of system programs. This feature is enabled by default. This
          feature consumes a certain amount of performance. If your performance
          is not enough, you can choose to disable it.
```

根据上面的描述，我们主要使用`--config`参数，该参数表示配置文件的位置。默认值是：`config.yaml`。

当`--config`为`config.yaml`（默认值）时，`--output-config`则会默认设置为`config.output.yaml`，并将配置文件输出到此位置。
输出的配置文件是完整版，包含全部选项和默认选项的，同时过滤非法选项。

### 配置文件
配置文件是`yaml`文件，请看以下配置文件：

```yaml
mode: debug  # 运行模式（Debug/Release/Test）
log-level: debug  # 日志记录登记
log-tag: enable  # 是否输出标签日志（Debug使用）
time-zone: Local  # 时区（UTC/Local/指定时区），若指定时区不存在，会退化到Local（本地电脑时区），若仍不存在则退化到UTC

tcp:  # TCP转啊规则
    rules:  # IP规则集
        - nation: ""  # 国家（精准）
          nation-vague: ""  # vague和不含vague的相比是模糊匹配, nation-vague设置为X，则可悲aX、bX、Xc、X、XX等模糊匹配
          province: ""  # 省份（规则同时）
          province-vague: ""
          city: ""  # 城市
          city-vague: ""
          isp: ""  # ISP
          isp-vague: ""
          # 上述信息均为地址信息，选填信息，当IP无法定位时，包含地址信息的规则会忽略
          
          ipv4: ""
          ipv6: ""
          ipv4cidr: 192.168.3.0/24
          ipv6cidr: ""
          # 上述为IP信息，四个最少选一个，若想表示全部IP，可为ipv4cidr选填0.0.0.0/0
          
          banned: disable  # 该规则效果：enable表示封禁，disable表示放行
          # 当以上条件和请求来访的ip一致（地区信息每一项为和关系，留空表示不启用，IP信息为或关系，满足一个即为命中规则。
          # 必须要IP信息和地址信息都命中规则才算命中，若无法获取IP的地址信息，则只能命中哪些没有地址信息的策略
          
    default-banned: disable  # 默认规则是否为banned：enable开启表示当上述规则均不匹配时拒绝该链接，disable表示默认放行
    always-allow-intranet: enable # 总是允许内网访问和本地回环（不需要上述规则集检查，但需要查看数据库是否封禁该IP）
    always-allow-loopback: enable # 总是允许本地回环访问（不需要上述规则集检查，也不需要经过数据库）
    
    # START 此处为一组 若 interface-name 留空则该组的配置不生效，网络性能监控器不使用
    interface-name: 以太网  # 网卡名称
    data-collection-cycle-seconds: 5  # 数据收集周期：建议5s
    statistical-time-span-seconds: 1800 # 数据统计时间跨度，单位秒（计算平均值时，时间的跨度。例如获取5分钟内接受到的数据包，然后除以5，得到每秒平均bytes，供下文使用
    statistical-period-seconds: 10 # 数据统计周期，多长时间进行一次数据统计，以及给出是否启用限流
    receive-bytes-of-cycle: 10kb # 入网流量限制（单位每秒）, 0 表示不i按照
    transmit-bytes-of-cycle: 10kb # 出网流量限制（单位每秒）, 0 表示不限制
    stop-accept-time-limit-seconds: 3600 # 高负荷多久后关停服务（单位：秒）
    
    forward:  # 转发规则
        - src: 8888  # 监听端口
          dest: localhost:8080  # 目标地址（域名可自动解析为ipv4和ipv6）
          ipv4-dest: ""  # 回源ipv4地址（权重比 dest 高）
          ipv6-dest: ""  # 回源ipv6地址 （权重比 dest 高）
          allow-cross: enable  # 允许交叉回原
          # 当你的服务器支持ipv4和ipv6，但只有ipv4或ipv6回源地址时，可以使用交叉功能，例如：让ipv6流量转发到ipv4。但是这种转发将不会使用Proxy协议。
          # 一般来说，启用了交叉，并设置了ipv4地址而没设置ipv6地址，则表示接收到ipv6信号要转发到ipv4
          # 但是若设置了dest，且从dest可以解析出ipv6，或ipv4地址也可以解析出ipv6，ipv6的流量将会转发上前述的ipv6地址时，前提是开启了交叉回源
          ipv4-src-proxy: enable  # ipv4监听时启动Proxy（启用后不影响接收非Proxy请求）
          ipv6-src-proxy: enable  # ipv6监听时启动Proxy（启用后不影响接收非Proxy请求）
          ipv4-dest-proxy: enable  # ipv4转发到目标地址时，是否启动Proxy。若是交叉回原，且为跨协议转发（例如 ipv4 转发到 ipv6）则忽略此处设定，均不使用Proxy协议
          ipv4-dest-proxy-version: 1 # ipv4转发到目标地址时使用的Proxy协议版本（截止至2025/2/16仅支持 1, 2），-1表示使用最新，0 表示使用默认（版本1）。尽当ipv4-dest-proxy启用时生效。
          ipv6-dest-proxy: enable # ipv4转发到目标地址时，是否启动Proxy。若是交叉回原，且为跨协议转发（例如 ipv4 转发到 ipv6）则忽略此处设定，均不使用Proxy协议
          ipv6-dest-proxy-version: 1 # ipv6转发到目标地址时使用的Proxy协议版本（截止至2025/2/16仅支持 1, 2），-1表示使用最新，0 表示使用默认（版本1）。尽当ipv6-dest-proxy启用时生效。

ssh:
    rules:  # 参照上文
        - nation: ""
          nation-vague: ""
          province: ""
          province-vague: ""
          city: ""
          city-vague: ""
          isp: ""
          isp-vague: ""
          ipv4: ""
          ipv6: ""
          ipv4cidr: 192.168.3.0/24
          ipv6cidr: ""
          banned: disable

    count-rules:  # 访问计数规则
        # 在规定时间（seconds）内，访问次数超过规定（try-count）次，则封禁规定时长（banned-second）。
        # 注意：try-count越大，seconds也要越大，并且较大者排在配置列表更前面
        # 此处是全局访问计数规则设定，每个转发服务可单独设定访问计数规则
        # 若转发服务未单独设定访问计数规则，则采用全局访问计数规则
        # 若全局访问计数规则也未设定，则采用默认值：3分钟内访问超过5次封禁10分钟。
        - try-count: 5
          seconds: 600
          banned-seconds: 1200
    
    default-banned: enable  # 默认规则是否为banned：enable开启表示当上述规则均不匹配时拒绝该链接，disable表示默认放行
    always-allow-intranet: disable # 总是允许内网访问和本地回环（不需要上述规则集检查，但需要查看数据库是否封禁该IP）
    always-allow-loopback: enable # 总是允许本地回环访问（不需要上述规则集检查，也不需要经过数据库）
    
    forward:  # 转发规则（见上文）
        - src: 8844
          dest: localhost:8844
          ipv4-dest: ""
          ipv6-dest: ""
          allow-cross: enable
          ipv4-src-proxy: enable
          ipv6-src-proxy: enable
          ipv4-dest-proxy: disable
          ipv4-dest-proxy-version: 1
          ipv6-dest-proxy: disable
          ipv6-dest-proxy-version: 1
          count-rules: [] # 可单独设定访问计数规则

api:
    app-code: # 阿里云市场 app-code
    # 需要调用的阿里云 云市场API
    #  1. IP定位：【无限免费】全球IP归属地查询-IP地址查询-IP城市查询-IP地址归属地-IP地址-IP地址查询-IP地址查询接口-ipv6
    #     API：https://kzipglobal.market.alicloudapi.com/api/ip/query
    #     云市场：https://market.aliyun.com/apimarket/detail/cmapi00066996?spm=5176.730005.result.2.38ae32e2LxRNzw&innerSource=search_ip%E5%BD%92%E5%B1%9E%E5%9C%B0#sku=yuncode6099600002

    webhook: # 企业微信机器人 Webhook，可为空，关闭企业微信推送

smtp:  # 发送邮件消息推送
    address: # smtp 服务器地址，可为空，为空表示关闭smtp
    user: # smtp 用户名（邮件），可为空，为空表示关闭smtp
    password: # smtp 用户密码
    recipient:
        - xxx@wxample.com  # 接收邮件通知的用户

redis:
    address: localhost:6379 # redis 服务器地址
    password: '123456' # redis 服务器密码
    db: 0 # redis 数据库

sqlite:
    path: data.db  # SQLite数据库位置
    active-close: disable  # 是否启用主动关闭数据库（一般情况下都不需要启用）
    clean: # 数据库清理
        execution-interval-hour: 6 # 数据库清理间隔时长（单位：小时）
        iface-record-save-retention-period: 3M # 网卡数据保留时长（3M：3个月）
        ssh-record-save-retention-period: 3M # SSH连接数据保留时长（3M：3个月）
```

## 构建与运行
### 构建
使用`go build`指令进行编译。
```shell
$ go build github.com/SongZihuan/huan-springboard/src/cmd/huanspringboard/hsbv1
```

生产环境下可以使用一些编译标志来压缩目标文件大小。
```shell
$ go build -trimpath -ldflags='-s -w' github.com/SongZihuan/huan-springboard/src/cmd/huanspringboard/hsbv1
```

### 运行
执行编译好的可执行文件即可。具体命令行参数可参见上文。

## 协议
本软件基于 [MIT LICENSE](/LICENSE) 发布。
了解更多关于 MIT LICENSE , 请 [点击此处](https://mit-license.song-zh.com) 。
## 协议
本软件基于 [MIT LICENSE](/LICENSE) 发布。
了解更多关于 MIT LICENSE , 请 [点击此处](https://mit-license.song-zh.com) 。
