# xmppbot

xmpp 聊天机器人。

## 模块说明

### 内置模块

sudo -- 超级管理员模块，提供--sudo开头的命令响应

### 扩展模块

bot  -- 好友模块，提供--bot开头的命令

room -- 聊天室模块，提供--room开头的命令，并自动响应聊天室消息

## 命令说明

### 好友命令

好友命令需要添加机器人为好友，并通过双向认证。

然后以聊天的形式发送。

超级管理员命令(需要以好友方式从管理员帐号发出):

  --sudo help
  --sudo list-all-plugins     列出所有的模块(管理员命令)
  --sudo list-plugins         列出当前启用的模块(管理员命令)
  --sudo disable <Plugin>     禁用某模块(管理员命令)
  --sudo enable <Plugin>      启用某模块(管理员命令)

  --bot help                  查看帮助
  --bot nick <nick name>      更改昵称
  --bot status <new status>   更新状态

  --room help                 查看帮助
  --room msg <msg>            让机器人在聊天室中发送消息msg
  --room nick <NickName>      修改机器人在聊天室的昵称为NickName
  --room block <who>          屏蔽who，对who发送的消息不响应
  --room unblock <who>        重新对who发送的消息进行响应

### 聊天室命令

聊天室命令需要从聊天室发出。

  --help              显示帮助信息
  --date              显示日期
  --blockme           停止抓取自己发送的链接标题
  --unblockme         恢复抓取自己发送的链接标题
  --ip                查询ip地址
  --weather <城市>    查询天气(未实现)
  --version           显示xmppdog版本信息
  --pkg    <pkg>      查询linux软件包


## 自动响应聊天室消息
