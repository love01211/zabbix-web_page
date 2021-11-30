# zabbix-agent web.page.*监控项在windows上无法正常使用问题解决

web.page.get[host,<path>,<port>]<br/>
web.page.perf[host,<path>,<port>]<br/>
web.page.regexp[host,<path>,<port>,regexp,<length>,<output>]<br/>
  
以上三个监控项在linux使用正常，windows使用报错为Unsupported item key.<br/>

## build<br/>
  go build web_page.go<br/>

## 使用
  1.拷贝文件到zabbix-agent的目录中<br/>
  2.配置文件zabbix-agent.conf中新增<br/>
  UserParameter=win_get_regexp[*],"C:\Program Files\zabbix_agent\web_page.exe" $1 $2 $3 $4 $5 $6 $7<br/>
  3.通过zabbix_get测试<br/>
  zabbix_get -s 192.168.113.198 -k win_get_regexp['web.page.regexp','https://www.cnblogs.com/lemon-le/p',"12750388.html","443","\d{3}","",""]<br/>
  200
  
## 补充
  使用方法和zabbix中的监控项基本一直，只不过第一个参数，决定是使用web.page.get/web.page.perf/web.page.regexp。
