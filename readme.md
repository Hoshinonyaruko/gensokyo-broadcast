<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/gensokyo-broadcast">
    <img src="images/head.gif" width="200" height="200" alt="gensokyo">
  </a>
</p>

<div align="center">

# gensokyo-broadcast

_✨ 基于 [OneBot](https://github.com/howmanybots/onebot/blob/master/README.md) Onebotv11广播命令行工具 ✨_  


</div>

<p align="center">
  <a href="https://raw.githubusercontent.com/hoshinonyaruko/gensokyo/main/LICENSE">
    <img src="https://img.shields.io/github/license/hoshinonyaruko/gensokyo-broadcast" alt="license">
  </a>
  <a href="https://github.com/hoshinonyaruko/gensokyo-broadcast/releases">
    <img src="https://img.shields.io/github/v/release/hoshinonyaruko/gensokyo?color=blueviolet&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/howmanybots/onebot/blob/master/README.md">
    <img src="https://img.shields.io/badge/OneBot-v11-blue?style=flat&logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABABAMAAABYR2ztAAAAIVBMVEUAAAAAAAADAwMHBwceHh4UFBQNDQ0ZGRkoKCgvLy8iIiLWSdWYAAAAAXRSTlMAQObYZgAAAQVJREFUSMftlM0RgjAQhV+0ATYK6i1Xb+iMd0qgBEqgBEuwBOxU2QDKsjvojQPvkJ/ZL5sXkgWrFirK4MibYUdE3OR2nEpuKz1/q8CdNxNQgthZCXYVLjyoDQftaKuniHHWRnPh2GCUetR2/9HsMAXyUT4/3UHwtQT2AggSCGKeSAsFnxBIOuAggdh3AKTL7pDuCyABcMb0aQP7aM4AnAbc/wHwA5D2wDHTTe56gIIOUA/4YYV2e1sg713PXdZJAuncdZMAGkAukU9OAn40O849+0ornPwT93rphWF0mgAbauUrEOthlX8Zu7P5A6kZyKCJy75hhw1Mgr9RAUvX7A3csGqZegEdniCx30c3agAAAABJRU5ErkJggg==" alt="gensokyo">
  </a>
  <a href="https://github.com/hoshinonyaruko/gensokyo-broadcast/actions">
    <img src="images/badge.svg" alt="action">
  </a>
  <a href="https://goreportcard.com/report/github.com/hoshinonyaruko/gensokyo-broadcast">
  <img src="https://goreportcard.com/badge/github.com/hoshinonyaruko/gensokyo-broadcast" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="https://github.com/howmanybots/onebot/blob/master/README.md">文档</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo-broadcast/releases">下载</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo-broadcast/releases">开始使用</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo-broadcast/blob/master/CONTRIBUTING.md">参与贡献</a>
</p>
<p align="center">
  <a href="https://gensokyo.bot">项目主页:gensokyo.bot</a>
</p>

## 介绍

gensokyo-broadcast是为onebotV11的正向http api设计的广播工具，对标nb和koishi内部插件提供的broadcast支持，为ob11标准的gensokyo及其他实现提供同等能力。

可根据相应平台的发送规则，所限制的频率，遵守内容规范的前提之下，发送机器人的维护信息，更新公告，活动公告。

提示，发送公告时，请遵守相关内容、频率规则，合理约束推送概率、范围，若不遵守相应规则产生的机器人被封禁，被停用后果，请自行承担。

特性，支持设置任意间隔时间，任意概率，会保存每一次发送后的结果，进度，可**断点续发**。可发送CQ码，从而可以组合图文，markdown等丰富类型信息。

可自行编辑发送txt列表，自定义广播范围。

欢迎issue提交功能建议！

## 帮助与支持
持续完善中.....交流群:196173384

欢迎测试,询问任何有关使用的问题,有问必答,有难必帮~

## WEBUI

![效果图](/pic/1.png)

![效果图](/pic/2.png)

WEBUI可配置登入密码,在手机和远程使用,可以选择载入推送模板,方便的进行模板推送.

## API

[API文档](/docs/api文档.md):可自行调用,将推送设计为指令\或自行编写UI\工具

## 兼容性与用法

安装

查看帮助的方法，gb.exe -h

下载release中的可执行文件,名称为gb-xxxx.exe，使用cmd，cd到当前exe路径，并输入可执行文件名运行。

为了方便使用，你可更改exe名称为简单名称，比如gb

或在文件浏览器中输入cmd运行，命令行程序不能直接双击exe运行。

设置HTTP API的地址，是gensokyo或onebot实现端的http 正向 api地址。


## 命令行参数说明

该工具支持以下命令行参数：

- `-a`：**必须**。设置OnebotV11 HTTP API的地址。示例：`-a http://localhost:8080`
- `-p`：**可选**。指定群列表的txt文件名（不包括.txt后缀）。示例：`-p group_list`，不填则自动获取并储存。
- `-w`：**必须**。要发送的信息内容。如果参数值包含`.txt`则尝试从对应的txt文件中读取内容，一行一条广播，否则直接将参数值作为消息内容。示例：`-w message.txt` 或 `-w '这是一条消息'||'这是另一条消息'`
- `-s`：**必须**。存档名，指定`-save`文件路径，用于断点续发。指定新文件名代表从头开始任务。不需要加`-save`和后缀。示例：`-s 本次任务代号`
- `-d`：**可选**。设置每条信息推送时间间隔（秒）。默认为10秒。示例：`-d 15`
- `-c`：**可选**。设置每个群推送的概率（百分比）。默认为100%，即总是推送。示例：`-c 50`
- `-h`：**可选**。显示帮助信息。不需要值，仅标志存在即可。

## 使用示例

### 发送固定消息到群组列表

假设你有一个名为`group_list`的文本文件，其中包含了群组的标识符，且想要发送文本`这是一条测试消息`到这些群组，你可以使用以下命令：

```sh
qf -a http://localhost:8080 -p group_list -w '这是一条测试消息' -s 测试任务
```

### 从文件读取消息内容

如果你想要从一个文本文件（比如`message.txt`）中读取消息内容发送，可以使用以下命令：

```sh
qf -a http://localhost:8080 -p group_list -w message.txt -s 测试任务
```

### 设置时间间隔和推送概率

如果你想要每15秒发送一条消息，并且每个群组的推送概率为50%，你可以使用以下命令：

```sh
qf -a http://localhost:8080 -p group_list -w '这是一条测试消息' -s 测试任务 -d 15 -c 50
```

## 关于 ISSUE

以下 ISSUE 会被直接关闭

- 提交 BUG 不使用 Template
- 询问已知问题
- 提问找不到重点
- 重复提问

> 请注意, 开发者并没有义务回复您的问题. 您应该具备基本的提问技巧。  
> 有关如何提问，请阅读[《提问的智慧》](https://github.com/ryanhanwu/How-To-Ask-Questions-The-Smart-Way/blob/main/README-zh_CN.md)

## 性能

1mb内存占用 端口错开可多开 稳定运行无报错 连续任务不崩溃中断 可断点续发