package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ipParse(ipRange string) ([]string, error) {
	parts := strings.Split(ipRange, ".")
	if len(parts) != 4 {
		return nil, fmt.Errorf("无效的ip地址或ip地址范围: %s", ipRange)
	}

	var start, end int
	var err error
	ipParts := make([][]string, 4)
	for i, part := range parts {
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("无效的ip地址或ip地址范围: %s", ipRange)
			}

			start, err = strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("无效的ip地址或ip地址范围: %s", ipRange)
			}

			end, err = strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("无效的ip地址或ip地址范围: %s", ipRange)
			}

			if start > end {
				return nil, fmt.Errorf("无效的ip地址或ip地址范围: %s", ipRange)
			}

			for j := start; j <= end; j++ {
				ipParts[i] = append(ipParts[i], strconv.Itoa(j))
			}
		} else {
			ipParts[i] = []string{part}
		}
	}

	var ips []string
	for i := 0; i < len(ipParts[0]); i++ {
		for j := 0; j < len(ipParts[1]); j++ {
			for k := 0; k < len(ipParts[2]); k++ {
				for l := 0; l < len(ipParts[3]); l++ {
					ip := fmt.Sprintf("%s.%s.%s.%s", ipParts[0][i], ipParts[1][j], ipParts[2][k], ipParts[3][l])
					ips = append(ips, ip)
				}
			}
		}
	}

	return ips, nil
}

func portParse(s string) []string {
	var ports []string
	for _, sub := range strings.Split(s, ",") {
		re, _ := regexp.Compile(`(\d+)-(\d+)`)
		matches := re.FindAllStringSubmatch(sub, -1)
		if len(matches) > 0 {
			var interval []string
			for _, match := range matches {
				start, _ := strconv.Atoi(match[1])
				end, _ := strconv.Atoi(match[2])
				for i := start; i <= end; i++ {
					interval = append(interval, strconv.Itoa(i))
				}
			}
			ports = append(ports, interval...)
		} else {
			ports = append(ports, sub)
		}
	}

	return ports
}

func portCheck(ip, port string) {
	_, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", ip, port), time.Second)
	if err == nil {
		fmt.Printf("%v开放端口%v\n", ip, port)
	}
}

func main() {
	help := `
端口扫描

%v ip port

  ip	要扫描的ip地址
        多个ip地址用英文减号(-)相连，例如：192.168.0.1-100

  port	要扫描的端口号
        多个端口可用空格( )或英文逗号(,)隔开，例如：80 81,82
        还可用英文减号(-)相连表示连续的端口号，例如：80-83

`
	args := os.Args
	if len(args) < 3 {
		fmt.Printf(help, args[0])
		os.Exit(1)
	}

	ipAddr := args[1]
	ipNet, _ := ipParse(ipAddr)

	var portList []string
	for i := 2; i < len(args); i++ {
		ports := portParse(args[i])
		portList = append(portList, ports...)
	}

	now := time.Now()
	fmt.Println("开始扫描:", now.Format(time.DateTime))

	var ipWg sync.WaitGroup
	for i := 0; i < len(ipNet); i++ {
		ip := ipNet[i]
		ipWg.Add(1)
		go func(ip string) {
			defer ipWg.Done()

			var portWg sync.WaitGroup
			for idx, port := range portList {
				portWg.Add(1)
				go func(port string) {
					defer portWg.Done()
					portCheck(ip, port)
				}(port)
				if idx >= len(portList)-1 || idx%100 == 0 {
					time.Sleep(time.Second)
					portWg.Wait()
				}
			}
		}(ip)
		if i >= len(ipNet)-1 || i%100 == 0 {
			time.Sleep(time.Second)
			ipWg.Wait()
		}
	}

	diff := time.Since(now)
	fmt.Println("扫描结束，耗时:", diff)
}
