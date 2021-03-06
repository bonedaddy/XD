package swarm

import (
	"net"
	"strconv"
	"time"
	"xd/lib/log"
	"xd/lib/sync"
	"xd/lib/tracker"
)

const DefaultAnnounceNumWant = 10
const DefaultAnnouncePort = 6881

type torrentAnnounce struct {
	access   sync.Mutex
	next     time.Time
	fails    time.Duration
	announce tracker.Announcer
	t        *Torrent
}

func (a *torrentAnnounce) tryAnnounce(ev tracker.Event) (err error) {
	a.access.Lock()
	if time.Now().After(a.next) {
		la := a.t.Network().Addr()
		if la.Network() == "i2p" {
		}
		req := &tracker.Request{
			Infohash:   a.t.st.Infohash(),
			PeerID:     a.t.id,
			Event:      ev,
			NumWant:    DefaultAnnounceNumWant,
			Left:       a.t.st.DownloadRemaining(),
			GetNetwork: a.t.Network,
		}
		if la.Network() == "i2p" {
			req.Port = DefaultAnnouncePort
		} else {
			var port string
			_, port, err = net.SplitHostPort(la.String())
			req.Port, err = strconv.Atoi(port)
			if err != nil {
				return
			}
		}
		if ev == tracker.Stopped {
			req.NumWant = 0
		}
		var resp *tracker.Response
		log.Infof("announcing to %s", a.announce.Name())
		resp, err = a.announce.Announce(req)
		backoff := a.fails * time.Minute
		a.next = resp.NextAnnounce.Add(backoff)
		if err == nil && ev != tracker.Stopped {
			a.t.addPeers(resp.Peers)
		}
	}
	a.access.Unlock()
	return
}
