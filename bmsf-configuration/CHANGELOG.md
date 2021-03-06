# CHANGELOG

## plugin-v2.0.11.6

1. 去除本地文件Effect
2. 增加Reset缓存支持
3. 更新sidecar环境变量文档
4. 去除无用代码

## v2.0.13.14

1. ProcAttr接口支持覆盖写
2. 批量创建ProcAttr默认覆盖原始非唯一索引信息

## plugin-v2.0.11.5

1. Agent消息Meta信息处理
2. 增加内部通道消息日志

## v2.0.13.13

1. Matched/Effected接口增加过期离线实例处理
2. Tunnel会话刷新优化
3. Tunnel通道消息追踪
4. Tunnel支持多连接
5. Tunnel支持消息熔断

## v2.0.13.8

1. 应用实例会话和版本生效信息同步优化
2. 下线实例History接口
3. 支持ProcAttr纳入effect查询结果


## plugin-v2.0.11.4

1. 应用实例会话和版本生效信息同步优化

## v2.0.13.5

1. 增加内部业务统计支持
2. 修复SQL内错误effectcode值
3. 新版本effect查询接口
4. 支持资源删除时同步策略

## v2.0.13.1

1. 修复本地CAS的策略即时同步问题

## plugin-v2.0.11.3

1. 新版本effect查询接口兼容错误码

## v2.0.13

1. 去除原信令链路, 完成链路合并
2. 同步打包流程去除无用模块文件
3. 更新安装脚本
4. 更细架构说明文档

## v2.0.12

1. 新版本Patcher纳入打包流程
2. 修复部分内置错误
3. 优化内部模块性能
4. 支持选主模式

## v2.0.11.3

1. patcher支持定时任务
2. 增加自动创建Cron文件逻辑
3. cron内部逻辑优化
4. 增加定时任务执行SEQ

## v2.0.11.2

1. 正式放开配置列表拉取过滤

## plugin-v2.0.11.2

1. 快速拉取日志优化稳定版

## v2.0.11.1

1. 配置列表拉取支持过滤
2. 更新下发通道拉取策略控制机制

## plugin-v2.0.11.1

1. 自动快速拉取优化

## v2.0.11

1. 配置类型扩展支持
2. 更新部分接口文档

## plugin-v2.0.11

1. 配置类型扩展版本

## v2.0.10.3

1. 支持批量创建进程绑定信息
2. 全局分页大小调整
3. 更新接口文档

## v2.0.10.2

1. 上报在线状态错误信息优化

## plugin-v2.0.10.2

1. 插件下沉会话服务分页调整

## v2.0.10.1

1. 默认下发频率控制调整

## plugin-v2.0.10.1

1. 上报在线状态错误信息优化

## v2.0.10

1. 上报信息优化

## v2.0.9

1. 策略控制匹配问题修复

## v2.0.8.15

1. 核心模块日志修整，减少过大日志输出
2. 调整模块Action调用日志级别
3. 修复控制器日志未能关联seq的问题

## v2.0.8.14

1. 优化网关代理默认网络配置参数
2. 去除无用公共函数

## v2.0.8.13

1. 修改默认全局配置项
2. 修改Processer默认数量

## v2.0.8.12

1. 节点拉取版本日志优化

## v2.0.8.11

1. 修复部分接口未正确纳入TX事务问题

## v2.0.8.10

1. 网关请求Metrics采集优化，去除非固定数据采集

## v2.0.8.9

1. 完善鉴权框架
2. 支持RBAC、ABAC、ACL混合鉴权模型

## v2.0.8.8

1. 修复ProcAttr列表查询问题
2. 修复分页查询总数检索的问题

## v2.0.8.7

1. 增加鉴权框架

## v2.0.8.6

1. 修复GCache会话剔除问题

## v2.0.8.5

1. 优化进程链路会话剔除逻辑

## v2.0.8.4

1. 修复gRPC-gateway关于PB转换64位整型问题
2. 增加PB开发指南，说明内部协议使用规范

## v2.0.8.3

1. 增加PB协议序列化公共函数
2. 修复网关因PB默认JSON标签未能正确返回相关字段的问题

## v2.0.8.2

1. 策略标签引擎支持多组模式
2. 修改相同接口统一标签策略规则

## v2.0.8.1

1. 增加healthz接口
2. 增加版本查询接口

## v2.0.8

1. 重构GSE通道协议

## v2.0.7.4

1. GSE下行客户端增加锁保护

## v2.0.7.3

1. 规范打包，增加企业版交包工具

## v2.0.7.2

1. 修复gorm适配器未知字段问题

## v2.0.6.0

1. 修复UUID问题
2. 修复制品库短名称问题

## v2.0.5.9

1. 支持附带内容和策略提交接口

## v2.0.5.8

1. 外吐部分参数到CMDLine
2. 增加Metrics相关配置到Flag

## v2.0.4.7

1. 增加Patcher模块

## v2.0.4.6

1. 修复procattr标签格式错误的问题

## v2.0.4.3

1. Tunnel信令协议优化

## v2.0.4.2

1. 修复SSL认证问题

## v2.0.3.10

1. 增加Etcd SSL配置支持
2. 补充部分模板操作审计修改相关逻辑错误码

## v2.0.3.8

1. multi commit/release事务性优化
2. 删除无用LRU缓存
3. template相关接口事务性优化
4. 针对multi场景支持强制cancel逻辑

## v2.0.3.4

1. 修改资源限制机制，更新API文档
