# API 调用文档

## 概述
该API接口允许用户通过HTTP请求触发程序，根据传入的参数执行不同的功能。以下是可用的参数列表及其详细说明。

请确保在实际应用中，对所有参数值进行urlencode以避免URL解析错误。

请求需要携带cookie,可从浏览器f12获取.token有效期1个月.文末有cookie获取和设置教程.

如果不提供cookie参数,将会得到{"error":"Unauthorized: Cookie not provided"}报错.

## 参数

### `-a` (HTTP API 的地址)
- **字段名**: `a`
- **类型**: `string`
- **描述**: 指定后端服务的HTTP API地址，用于网络请求。

### `-p` (群列表的文件名)
- **字段名**: `p`
- **类型**: `string`
- **描述**: 指定要读取的群列表txt文件的文件名(不包含.txt后缀)。

### `-w` (要发送的信息)
- **字段名**: `w`
- **类型**: `string`
- **描述**: 指定要发送的消息内容。如果内容包含`.txt`后缀，则尝试从对应的txt文件中读取内容。

### `-d` (每条信息推送时间的间隔)
- **字段名**: `d`
- **类型**: `int`
- **默认值**: `10`
- **描述**: 设置每条消息的推送时间间隔（单位：秒）。

### `-c` (每个群推送的概率)
- **字段名**: `c`
- **类型**: `int`
- **默认值**: `100`
- **描述**: 每个群组推送消息的概率（单位：百分比%）。

### `-h` (显示帮助信息)
- **字段名**: `h`
- **类型**: `bool`
- **默认值**: `false`
- **描述**: 当此参数被设置时，程序将输出帮助信息并退出。

### `-s` (读取-save文件路径)
- **字段名**: `s`
- **类型**: `string`
- **描述**: 指定用于断点续传的-save文件的路径。

### `-g` (gensokyo过滤子频道)
- **字段名**: `g`
- **类型**: `bool`
- **默认值**: `false`
- **描述**: 是否启用gensokyo子频道的过滤。

### `-f` (私聊模式)
- **字段名**: `f`
- **类型**: `bool`
- **默认值**: `false`
- **描述**: 如果设置为true，则模式限定为私聊，不发送群广播信息。

### `-t` (access_token)
- **字段名**: `t`
- **类型**: `string`
- **描述**: 如果设置了HTTP的密钥，则需要此参数来进行验证。

### `-r` (打乱群/好友列表顺序)
- **字段名**: `r`
- **类型**: `bool`
- **默认值**: `false`
- **描述**: 是否打乱群组和好友列表的顺序。

## 示例调用

通过curl发送带参数的请求示例：

```bash
curl "http://localhost:60123/run?p=group_list&w=这是一条消息&d=15&a=http://example.com&c=80&s=savepath&g=true&f=true&t=your_token&r=true"
```

### 获取Cookie

1. 打开浏览器，导航到您的网站。
2. 使用F12键打开开发者工具。
3. 切换到“网络(Network)”标签页。
4. 在网站上触发一个请求（比如登录或加载页面）。
5. 查看网络请求列表，找到任一已发送的请求，点击它。
6. 在请求详情中找到“请求头(Request Headers)”部分，复制`Cookie`字段的内容。

### Python API调用示例

这个示例展示了如何使用Python `requests`库发送一个带有cookie的HTTP GET请求：

```python
import requests

# API的URL
url = 'https://example.com/api/run'

# 需要发送的参数
params = {
    'a': 'http://127.0.0.1:42001',
    'p': '1716208337-5月20日babyq2',
    'w': 'baby.txt',
    'd': '10',
    'c': '100',
    's': '5月20日babyq2',
    'g': 'true',
    't': 'xyy520499'
}

# 从浏览器获取的cookie字符串
cookies = {
    'session_token': 'your_copied_cookie_here'  # 替换成从浏览器复制的cookie值
}

# 发送请求
response = requests.get(url, params=params, cookies=cookies)

# 输出响应内容
print(response.text)
```