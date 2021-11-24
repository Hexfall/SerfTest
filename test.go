package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/serf/serf"
	"log"
	"time"
)

var ip 	    = flag.String("ip", "127.0.0.1", "Own IP")
var port    = flag.Int("port", 9080, "Port")
var target  = flag.String("target", "", "Target")

var events = make(chan serf.Event)

func main() {
	flag.Parse()
	go func() {
		for {
			event := <-events
			log.Printf("Received event: %v\n", event)
		}
	}()

	config := getConfig(*ip, *port)
	client, err := serf.Create(config)
	if err != nil {
		log.Fatalf("Failed to create Serf Client. %v", err)
	}
	defer client.Shutdown()

	// If target is not self.
	if *target != "" {
		memberCount, err := client.Join([]string{ fmt.Sprintf("node-%s/%s", *target, *target) }, false)
		if err != nil {
			log.Fatalf("Failed to join ip. %v", err)
		} else {
			log.Printf("Successfully joined. %d buddies found!", memberCount)
		}
		for _, mem := range client.Members() {
			log.Printf("Found member: %s\n", mem.Name)
		}
	} else {
		log.Printf("Waiting for other connections on %s/%s", config.NodeName, config.MemberlistConfig.BindAddr)
	}

	// Monitor for joining members.
	for {
		time.Sleep(time.Second*10)
		log.Printf("Currently have %d buddies.\n", client.NumNodes() - 1)
		for _, member := range client.Members() {
			log.Printf("\t%s | %s:%d\n", member.Name, member.Addr, member.Port)
		}
	}
}

func getConfig(ip string, port int) *serf.Config {
	conf := serf.DefaultConfig()
	conf.Init()

	conf.MemberlistConfig.BindAddr = fmt.Sprintf("%s:%d", ip, port)
	conf.MemberlistConfig.BindPort = port
	conf.NodeName = fmt.Sprintf("node-%s:%d", ip, port)
	conf.ValidateNodeNames = false

	// Set probe intervals that are aggressive for finding bad nodes
	conf.MemberlistConfig.GossipInterval = 500 * time.Millisecond
	conf.MemberlistConfig.ProbeInterval = 5000 * time.Millisecond
	conf.MemberlistConfig.ProbeTimeout = 2500 * time.Millisecond
	conf.MemberlistConfig.TCPTimeout = 10000 * time.Millisecond
	conf.MemberlistConfig.SuspicionMult = 1

	// Set a short reap interval so that it can run during the test
	conf.ReapInterval = 1 * time.Second

	// Set a short reconnect interval so that it can run a lot during tests
	conf.ReconnectInterval = 100 * time.Millisecond

	// Set basically zero on the reconnect/tombstone timeouts so that
	// they're removed on the first ReapInterval.
	conf.ReconnectTimeout = 1 * time.Second
	conf.TombstoneTimeout = 1 * time.Second

	conf.EventCh = events

	return conf
}
