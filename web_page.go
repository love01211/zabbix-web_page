package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main()  {
	args_key := os.Args[1]
	args_params := os.Args[2:]

	//result, err := Export("", []string{"shenq.cc"})
	result, err := Export(args_key, args_params)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}

//func (p *Plugin) Export(key string, params []string, ctx plugin.ContextProvider) (interface{}, error) {
func Export(key string, params []string) (interface{}, error) {
	if len(params) == 0 || params[0] == "" {
		return nil, fmt.Errorf("Invalid first parameter.")
	}

	u, err := url.Parse(params[0])
	if err != nil {
		return nil, fmt.Errorf("Cannot parse url: %s", err)
	}

	if u.Scheme == "" || u.Opaque != "" {
		params[0] = "http://" + params[0]
	}

	if len(params) > 2 && params[2] != "" {
		params[0] += ":" + params[2]
	}

	if len(params) > 1 && params[1] != "" {
		if params[1][0] != '/' {
			params[0] += "/"
		}

		params[0] += params[1]
	}

	switch key {
	case "web.page.regexp":
		var length *int
		var output string

		if len(params) > 6 {
			return nil, fmt.Errorf("Too many parameters.")
		}

		if len(params) < 4 {
			return nil, fmt.Errorf("Invalid number of parameters.")
		}

		rx, err := regexp.Compile(params[3])
		if err != nil {
			return nil, fmt.Errorf("Invalid forth parameter: %s", err)
		}

		if len(params) > 4 && params[4] != "" {
			if n, err := strconv.Atoi(params[4]); err != nil {
				return nil, fmt.Errorf("Invalid fifth parameter: %s", err)
			} else {
				length = &n
			}
		}

		if len(params) > 5 && params[5] != "" {
			output = params[5]
		} else {
			output = "\\0"
		}

		//s, err := Get(params[0], time.Duration(p.options.Timeout)*time.Second, true)
		s, err := Get(params[0], time.Duration(15)*time.Second, true)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(strings.NewReader(s))
		for scanner.Scan() {
			if out, ok := ExecuteRegex(scanner.Bytes(), rx, []byte(output)); ok {
				if length != nil {
					out = CutAfterN(out, *length)
				}
				return out, nil
			}
		}

		return "", nil
	case "web.page.perf":
		if len(params) > 3 {
			return nil, fmt.Errorf("Too many parameters.")
		}

		start := time.Now()

		//_, err := web.Get(params[0], time.Duration(p.options.Timeout)*time.Second, false)
		//手动设定超时请求为30s
		_, err := Get(params[0], time.Duration(15)*time.Second, false)
		if err != nil {
			return nil, err
		}

		return time.Since(start).Seconds(), nil
	default:
		if len(params) > 3 {
			return nil, fmt.Errorf("Too many parameters.")
		}

		//return web.Get(params[0], time.Duration(p.options.Timeout)*time.Second, true)
		//手动设定超时请求为30s
		return Get(params[0], time.Duration(15)*time.Second, true)
	}

}

//新增GET函数
//https://github.com/zabbix/zabbix/blob/master/src/go/pkg/web/web.go
func Get(url string, timeout time.Duration, dump bool) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("Cannot create new request: %s", err)
	}

	req.Header = map[string][]string{
		//"User-Agent": {"Zabbix " + version.Long()},
		//手动写死
		"User-Agent": {"Zabbix 5.0"},
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
			//看着像是获取文件里的ip，注释
			//DialContext: (&net.Dialer{
			//	LocalAddr: &net.TCPAddr{IP: net.ParseIP(agent.Options.SourceIP), Port: 0},
			//}).DialContext,
		},
		Timeout:       timeout,
		CheckRedirect: disableRedirect,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Cannot get content of web page: %s", err)
	}

	defer resp.Body.Close()

	if !dump {
		return "", nil
	}

	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return "", fmt.Errorf("Cannot get content of web page: %s", err)
	}

	return string(bytes.TrimRight(b, "\r\n")), nil
}

func disableRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

//新增正则解析函数
//https://github.com/zabbix/zabbix/blob/master/src/go/pkg/zbxregexp/zbxregexp.go
func ExecuteRegex(line []byte, rx *regexp.Regexp, output []byte) (result string, match bool) {
	matches := rx.FindSubmatchIndex(line)
	if len(matches) == 0 {
		return "", false
	}
	if len(output) == 0 {
		return string(line), true
	}

	buf := &bytes.Buffer{}
	for len(output) > 0 {
		pos := bytes.Index(output, []byte{'\\'})
		if pos == -1 || pos == len(output)-1 {
			break
		}
		_, _ = buf.Write(output[:pos])
		switch output[pos+1] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i := output[pos+1] - '0'
			if len(matches) >= int(i)*2+2 {
				if matches[i*2] != -1 {
					_, _ = buf.Write(line[matches[i*2]:matches[i*2+1]])
				}
			}
			pos++
		case '@':
			_, _ = buf.Write(line[matches[0]:matches[1]])
			pos++
		case '\\':
			_ = buf.WriteByte('\\')
			pos++
		default:
			_ = buf.WriteByte('\\')
		}
		output = output[pos+1:]
	}
	_, _ = buf.Write(output)
	return buf.String(), true
}

//新增CutAfterN
//https://github.com/zabbix/zabbix/blob/master/src/go/internal/agent/options.go
func CutAfterN(s string, n int) string {
	var i int
	for pos := range s {
		if i >= n {
			return s[:pos]
		}
		i++
	}

	return s
}