package main

import (
	"fmt"
	"github.com/hashicorp/serf/serf"
	"os"
	"strconv"
	"time"
)

var eventCh = make(chan serf.Event)

func getConfig(ip string, port int) *serf.Config {
	conf := serf.DefaultConfig()
	conf.Init()

	conf.MemberlistConfig.BindAddr = ip + ":" + string(port)
	conf.MemberlistConfig.BindPort = port
	conf.NodeName = conf.MemberlistConfig.BindAddr

	// Set probe intervals that are aggressive for finding bad nodes
	conf.MemberlistConfig.GossipInterval = 50 * time.Millisecond
	conf.MemberlistConfig.ProbeInterval = 500 * time.Millisecond
	conf.MemberlistConfig.ProbeTimeout = 250 * time.Millisecond
	conf.MemberlistConfig.TCPTimeout = 1000 * time.Millisecond
	conf.MemberlistConfig.SuspicionMult = 1

	// Activate the strictest version of memberlist validation to ensure
	// we properly pass node names through the serf layer.
	conf.MemberlistConfig.RequireNodeNames = true

	// Set a short reap interval so that it can run during the test
	conf.ReapInterval = 1 * time.Second

	// Set a short reconnect interval so that it can run a lot during tests
	conf.ReconnectInterval = 100 * time.Millisecond

	// Set basically zero on the reconnect/tombstone timeouts so that
	// they're removed on the first ReapInterval.
	conf.ReconnectTimeout = 1 * time.Microsecond
	conf.TombstoneTimeout = 1 * time.Microsecond

	return conf
}

func main() {
	port, _ := strconv.Atoi(os.Args[2])
	SerfClient, _ := serf.Create(getConfig(os.Args[1], port))
	defer SerfClient.Shutdown()
	go eventHandler(&eventCh)
	if len(os.Args) > 3 {
		buddies, _ := SerfClient.Join([]string{ fmt.Sprintf("%s/%s", os.Args[3], os.Args[3]) }, false)
		fmt.Printf("%d buddies found!", &buddies)
	}

	for {
		time.Sleep(time.Second*4)
		fmt.Printf("Currently have %d buddies.\n", len(SerfClient.Members()) - 1)
		for _, member := range SerfClient.Members() {
			fmt.Println("\t" + member.Name)
		}
	}
}

func eventHandler(ch *chan serf.Event) {
	for {
		event := <- *ch
		fmt.Print("EVENT RECEIVED: ")
		fmt.Println(event.EventType())
	}
}
