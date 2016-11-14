package network

import (
	"fmt"
	"time"

	"monitor/alert"
	"monitor/format"
	"monitor/parse"

	"github.com/kevinchen/logx"
	"github.com/shirou/gopsutil/net"
)

type Network struct {
	Name       string
	NetTotalTx uint64
	NetTotalRx uint64
	TxSec      float64
	RxSec      float64
}

type Networks struct {
	Nets []Network
}

func newNetwork() *Networks {
	return &Networks{}
}

func GetNetwork() (*Networks, error) {
	network := newNetwork()
	return network.getNetwork()
}

func (n *Networks) getNetwork() (*Networks, error) {
	io, err := net.IOCounters(false)

	if err != nil {
		logx.FError("getNetWork to get net io counter error: %v", err)
		return nil, err
	}

	for _, v := range io {
		if v.Name == "all" {
			n.Nets = append(n.Nets, Network{
				Name:       v.Name,
				NetTotalTx: v.BytesSent,
				NetTotalRx: v.BytesRecv,
			})
		}
	}
	return n, nil
}

func MonitorNetwork() {

	go func() {
		network := newNetwork()
		logx.FInfo("%v", "Network module started!")
		if err := network.monitorNetwork(); err != nil {
			logx.FCritical("MonitorNetwork error: %v", err)
		}
	}()
}

func (n *Networks) monitorNetwork() error {

	cf := parse.NewCf()
	if cf.NetConfig == nil {
		logx.FError("%v", "NetConfig is nil!")
		return fmt.Errorf("NetConfig is nil!")
	}

	n.reset()
	t1 := time.NewTicker(time.Millisecond * time.Duration(cf.NetConfig.Duration))
	var rxpre uint64
	var txpre uint64
	for {

		networks, err := n.getNetwork()
		if err != nil {
			logx.FError("monitorNetwork error: %v", err)
			return err
		}

		for _, v := range networks.Nets {
			rxpre = v.NetTotalRx // receive
			txpre = v.NetTotalTx // send

			logx.FDebug("All receive data: %s", format.Translate(float64(rxpre)))
			logx.FDebug("All send data: %s", format.Translate(float64(txpre)))
		}

		//Clean slice.
		n.reset()
		select {
		case <-t1.C:
			networks, err := n.getNetwork()
			if err != nil {
				continue
			}

			for _, v := range networks.Nets {
				rxSpeed := float64(v.NetTotalRx-rxpre) / 1024
				txSpeed := float64(v.NetTotalTx-txpre) / 1024

				v.RxSec = rxSpeed
				v.TxSec = txSpeed
				logx.FDebug("Network speed  RX: %v%s, TX: %v%s", rxSpeed, "k/s", txSpeed, "k/s")
				// Uints is k/s
				if rxSpeed >= float64(cf.NetConfig.MaxRx) || txSpeed >= float64(cf.NetConfig.MaxTx) {
					content := fmt.Sprintf("Alarm: network speed beycond the threshold RX: %d%s, TX: %d%s! \nfrom: %v", (v.NetTotalRx-rxpre)/1000, "k/s", (v.NetTotalTx-txpre)/1000, "k/s", cf.Addr)
					header := "Network being monitored has exceeded the maximum!"
					timestamp := time.Now()
					alert.AlertConvergence(cf.NetConfig.EmailArray, timestamp, content, header)
				}
			}
			//Clean slice.
			n.reset()
		}
	}
	return nil
}

func (n *Networks) reset() {
	n.Nets = n.Nets[:0:0]
}
