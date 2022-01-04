# XJTUPortal
> A web portal authentication manager for XJTU iHarbor campus network

---

## 功能特色
* 简化登录操作，运行即可自动登录，一气呵成！
  > 需在配置文件设置运行模式为```command```且事先设置账号密码
* 全自动无人值守登录，设备超限时可按照策略自动下线设备，远程桌面 & 服务器玩家的福音！
  > 依赖外部支持，例如任务计划程序（Windows）或 crontab（Linux）
* 摆脱传统的 Web 页面操作，无视系统代理，有效避免因浏览器问题或系统代理问题而无法登录的问题！
* 上不了网却不知道为什么？不用担心！完善的日志记录功能帮助运维人员快速解决你的问题！
* 附带简单的网络诊断功能，可帮你初步判断网络故障，上网不求人！

## 快速上手
* 下载适用于当前系统和 CPU 架构的程序，拷贝至任意目录
  > Intel 和 AMD 所出品的绝大部分 CPU 的架构为 amd64  
  > 使用 ARM 架构 CPU 的消费级产品多为智能手机
* 确保程序和```config```文件夹位于同一目录下，```config```文件夹内的所有文件不要更改其文件名
* 运行程序（默认启用```interact```模式）
* 按照提示，选择“快速设置”，按照提示输入信息并确认
* 重新运行程序，之后可按照提示执行登录、登出等操作

## 使用技巧
* ```user-setting.yaml```中设置运行模式为```command```时，运行程序后将会直接执行登录操作，在```interact```模式下会进入交互界面
* 默认情况下，程序将会在其所在的目录下寻找 config 文件夹，可通过```-c```参数指定其它配置文件目录
  > ```-c```为程序运行唯一必要参数，除非程序所在目录包含 config 文件夹时可省略
* 请求运维人员诊断时，请先执行“临时切换日志级别为 DEBUG”操作
  > 默认的日志文件路径为```run.log```，请提供日志文本，**不要截图！更不要拍照！**
* 全自动无人值守登录示例
  > 使用 crontab 设置每 5 分钟检查一次网络状态，若下线则自动登录：
  > ```*/5 * * * * /usr/local/bin/xjtuportal -c /usr/local/etc/xjtuportal```  
  > 其中，```/usr/local/bin/xjtuportal```为程序所在目录，```/usr/local/etc/xjtuportal```为配置文件所在目录，请按照实际情况自行替换
## 注意事项
* 可通过参数```-h```获取运行参数设置帮助
* 更多功能配置请参考配置文件
