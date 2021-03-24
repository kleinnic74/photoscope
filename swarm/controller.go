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

type Peer struct {
	ID   string   `json:"id"`
	URL  string   `json:"url"`
	Type string   `json:"service"`
	Text []string `json:"text,omitempty"`
}

type Controller struct {
	instance string
	resolver *zeroconf.Resolver
	done     chan struct{}

	peers    map[string]Peer
	services map[string]string
	peerLock sync.Mutex
}

func NewController(instance string) (*Controller, error) {
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

func (c *Controller) ListenAndServe(ctx context.Context) {
	logger, ctx := logging.SubFrom(ctx, "swarm.controller")
	server, err := zeroconf.Register(c.instance, "_photoscope._tcp", "local.", 8080, []string{}, nil)
	if err != nil {
		logger.Error("Failed to publish zeroconf service: %s", zap.Error(err))
	}
	defer server.Shutdown()

	logger.Info("Looking for peers...")
	peerCh := make(chan *zeroconf.ServiceEntry)
	browseCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	c.resolver.Browse(browseCtx, "_services._dns-sd._udp", "", peerCh)

	for {
		select {
		case p := <-peerCh:
			if p == nil {
				continue
			}
			if p.Service == "_services._dns-sd._udp" {
				// DNS-SD service, query further
				service := p.Instance
				if strings.HasSuffix(service, ".") {
					service = service[:len(service)-1]
				}
				service = strings.TrimSuffix(service, p.Domain)
				// TODO: protect with mutex
				c.services[service] = service
				logger.Info("Received DNS-SD service",
					zap.String("peer.service", service),
					zap.String("peer.domain", p.Domain))
				c.resolver.Browse(ctx, service, "", peerCh)
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
		ID:   p.Instance,
		Type: p.Service,
		URL:  asURL(p),
		Text: p.Text,
	}
	if _, found := c.peers[peer.ID]; !found {
		c.peers[peer.ID] = peer
		logging.From(ctx).Info("Peer detected",
			zap.String("peer.instance", peer.ID),
			zap.String("peer.URL", peer.URL),
			zap.String("peer.type", peer.Type),
			zap.String("peer.hostname", p.HostName),
			zap.Strings("text", peer.Text))
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

func (c *Controller) GetPeers() (r []Peer) {
	c.peerLock.Lock()
	defer c.peerLock.Unlock()

	for _, p := range c.peers {
		r = append(r, p)
	}
	return
}

func (c *Controller) GetSeenServices() (r []string) {
	c.peerLock.Lock()
	defer c.peerLock.Unlock()

	for _, s := range c.services {
		r = append(r, s)
	}
	return
}
