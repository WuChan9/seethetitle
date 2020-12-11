# seethetitle

```
Usage: seethetitle.exe -n 192.168.0.1/24 -p 80,443

  -n, --net            扫描网段(default:127.0.0.1/32)
  -p, --port           扫描端口(default:80)
  --timeout            请求超时(default:3s)
  -t, --thread         并发数
```



## golang html 编码识别

> https://html.spec.whatwg.org/multipage/parsing.html#determining-the-character-encoding

　　2 遍的预扫描机制



　　charset.DetermineEncoding 通过读取前 1024 字节

```
_, charsetName, _ := charset.DetermineEncoding(body, "utf-8")
//windows-1252 可能是
if charsetName == "gbk" {
	body, err = simplifiedchinese.GBK.NewDecoder().Bytes(body)
	if err != nil {
		continue
	}
} 
```

