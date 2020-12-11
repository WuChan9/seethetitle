# seethetitle

```
Usage: seethetitle.exe -n 192.168.0.1/24 -p 80,443

  -n, --net            扫描网段(default:127.0.0.1/32)
  -p, --port           扫描端口(default:80)
  --timeout            请求超时(default:3s)
  -t, --thread         并发数
```



### html 编码识别

> https://html.spec.whatwg.org/multipage/parsing.html#determining-the-character-encoding

```
_, charsetName, _ := charset.DetermineEncoding(body, "")
```

