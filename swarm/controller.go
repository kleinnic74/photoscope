package swarm

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"bitbucket.org/kleinnic74/photos/logging"
	"github.com/grandcat/zeroconf"
	"go.uber.org/zap"
)

const (
	PhotoscopeSVCName = "_photoscope._tcp"
	DNSSDSVCName      = "_services._dns-sd._udp"
)

type Peer struct {
	Name string   `json:"name"`
	ID   string   `json:"id"`
	URL  string   `json:"url"`
	Type string   `json:"service"`
	Text []string `json:"text,omitempty"`
}

type PeerHandler func(context.Context, Peer)

type Controller struct {
	instance *Instance
	resolver *zeroconf.Resolver
	done     chan struct{}

	peers    map[string]Peer
	services map[string]string
	peerLock sync.RWMutex

	handlers []PeerHandler
}

var interestingServices = map[string]struct{}{
	PhotoscopeSVCName: {},
}

func NewController(instance *Instance) (*Controller, error) {
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		return nil, err
	}
	return &Controller{
		instance: instance,
		resolver: resolver,
		peers:    make(map[string]Peer),
		services: make(map[string]string),
		done:     make(chan struct{}),
	}, nil
}

func (c *Controller) OnPeerDetected(h PeerHandler) {
	c.handlers = append(c.handlers, h)
}

func (c *Controller) ListenAndServe(ctx context.Context) {
	logger, ctx := logging.SubFrom(ctx, "swarm.controller")
	server, err := zeroconf.Register(c.instance.Name, PhotoscopeSVCName, "local.", 8080, []string{fmt.Sprintf("id=%s", c.instance.ID)}, nil)
	if err != nil {
		logger.Error("Failed to publish zeroconf service: %s", zap.Error(err))
	}
	defer server.Shutdown()

	logger.Info("Looking for peers...")
	peerCh := make(chan *zeroconf.ServiceEntry)
	dnssdCh := make(chan *zeroconf.ServiceEntry)

	browseCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.resolver.Browse(browseCtx, DNSSDSVCName, "", dnssdCh)

	for {
		select {
		case p := <-dnssdCh:
			if p == nil {
				continue
			}
			// DNS-SD service, query further
			service := canonicalServiceName(p.Instance, p.Domain)
			logger.Info("Received DNS-SD service",
				zap.String("peer.service", service),
				zap.String("peer.domain", p.Domain))

			if _, found := interestingServices[service]; found {
				func() {
					c.peerLock.Lock()
					defer c.peerLock.Unlock()
					logger.Debug("Querying service instances", zap.String("peer.service", service))

					c.services[service] = service
				}()
				c.resolver.Browse(ctx, service, p.Domain, peerCh)
			}
		case p := <-peerCh:
			if p == nil {
				continue
			}
			c.peerDiscovered(ctx, p)
		case <-c.done:
			logger.Info("Shutting down")
			return
		}
	}
}

func (c *Controller) Shutdown() {
	close(c.done)
}

func (c *Controller) peerDiscovered(ctx context.Context, p *zeroconf.ServiceEntry) {
	c.peerLock.Lock()
	defer c.peerLock.Unlock()

	peer := Peer{
		Name: p.Instance,
		ID:   findID(p.Text),
		Type: p.Service,
		URL:  asURL(p),
		Text: p.Text,
	}
	if _, found := c.peers[peer.Name]; !found {
		c.peers[peer.Name] = peer
		logging.From(ctx).Info("Peer detected",
			zap.String("peer.instance", peer.Name),
			zap.String("peer.URL", peer.URL),
			zap.String("peer.type", peer.Type),
			zap.String("peer.hostname", p.HostName),
			zap.Strings("text", peer.Text))
		for _, h := range c.handlers {
			h(ctx, peer)
		}
	}
}

func asURL(p *zeroconf.ServiceEntry) string {
	if len(p.AddrIPv4) > 0 {
		return fmt.Sprintf("http://%s:%d", p.AddrIPv4[0], p.Port)
	}
	if len(p.AddrIPv6) > 0 {
		return fmt.Sprintf("http://%s:%d", p.AddrIPv6[0], p.Port)
	}
	return ""
}

func findID(txt []string) string {
	for _, t := range txt {
		kv := strings.SplitN(t, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "id" {
			return kv[1]
		}
	}
	return ""
}

func canonicalServiceName(name, domain string) string {
	if name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}
	name = strings.TrimSuffix(name, domain)
	if name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}
	return name
}

func (c *Controller) GetPeers() (r []Peer) {
	c.peerLock.RLock()
	defer c.peerLock.RUnlock()

	for _, p := range c.peers {
		r = append(r, p)
	}
	return
}

func (c *Controller) GetSeenServices() (r []string) {
	c.peerLock.RLock()
	defer c.peerLock.RUnlock()

	for _, s := range c.services {
		r = append(r, s)
	}
	return
}