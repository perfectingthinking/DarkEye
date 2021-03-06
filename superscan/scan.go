package main

import (
	"github.com/fatih/color"
	"github.com/zsdevX/DarkEye/common"
	"github.com/zsdevX/DarkEye/superscan/plugins"
	"strconv"
	"sync"
	"time"
)

//New add comment
func New(ip string) *Scan {
	return &Scan{
		Ip:                     ip,
		ActivePort:             "80",
		PortRange:              common.PortList,
		Callback:               callback,
		BarCallback:            barCallback,
		ThreadNumber:           200,
		BarDescriptionCallback: descCallback,
	}
}

//Run add comment
func (s *Scan) Run() {
	wg0 := sync.WaitGroup{}
	wg0.Add(1)
	go func() {
		s.preCheck()
		wg0.Done()
	}()
	defer wg0.Wait()

	fromTo, tot := common.GetPortRange(s.PortRange)
	taskAlloc := make(chan int, s.ThreadNumber)
	wg := sync.WaitGroup{}
	wg.Add(tot)
	for _, p := range fromTo {
		for p.From <= p.To {
			taskAlloc <- 1
			go func(port int) {
				for {
					s.Check(port)
					if !s.isFireWallNotForbidden() {
						//被防火墙策略限制探测，等待恢复期（恢复期比较傻，需要优化）。
						time.Sleep(time.Second * 10)
						//恢复后从中断的端口重新检测
						continue
					}
					break
				}
				<-taskAlloc
				wg.Done()
			}(p.From)
			p.From++
		}
	}
	wg.Wait()
}

//Check add comment
func (s *Scan) Check(p int) {
	defer func() {
		s.BarCallback(1)
	}()

	if s.ActivePort != "0" {
		if common.IsAlive(s.Ip, strconv.Itoa(p), s.TimeOut) == common.Alive {
			color.Green("\n%s %s:%s %v\n", "[√]",
				s.Ip, strconv.Itoa(p), "Opened")
		}
		//开启防火墙检测仅判断端口，不爆破
		return
	}

	plg := plugins.Plugins{
		TargetIp:     s.Ip,
		TargetPort:   strconv.Itoa(p),
		TimeOut:      s.TimeOut,
		PortOpened:   false,
		NoTrust:      s.NoTrust,
		Worker:       s.PluginWorker,
		DescCallback: s.BarDescriptionCallback,
	}
	plg.Check()
	if !plg.PortOpened {
		return
	}
	s.Callback(&plg)
}

func (s *Scan) preCheck() {
	if s.ActivePort != "0" {
		//开启防火墙检测仅判断端口，不探测
		return
	}

	plg := plugins.Plugins{
		TargetIp:     s.Ip,
		TimeOut:      s.TimeOut,
		DescCallback: s.BarDescriptionCallback,
	}
	plg.PreCheck()
	if len(plg.Cracked) == 0 {
		return
	}
	plg.TargetPort = "-"
	plg.TargetProtocol = "-"
	s.Callback(&plg)
}

func (s *Scan) isFireWallNotForbidden() bool {
	//为0不矫正
	if s.ActivePort == "0" {
		return true
	}
	maxRetries := 3
	for maxRetries > 0 {
		if common.IsAlive(s.Ip, s.ActivePort, s.TimeOut) == common.Alive {
			return true
		}
		maxRetries--
	}
	return false
}

func callback(a interface{}) {
	//todo
}

func barCallback(i int) {
	//todo
}

func descCallback(i string) {
	//todo
}
