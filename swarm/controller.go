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
	Name       string            `json:"name"`
	ID         InstanceID        `json:"id"`
	URL        string            `json:"url"`
	Type       string            `json:"service"`
	Properties map[string]string `json:"properties,omitempty"`
	IsSelf     bool              `json:"-"`
}

func propertiesAsTXT(p map[string]string) (txt []string) {
	for k, v := range p {
		txt = append(txt, fmt.Sprintf("%s=%s", k, v))
	}
	return
}

func propertiesFromTXT(txt []string) (p map[string]string) {
	p = make(map[string]string)
	for _, kv := range txt {
		parts := strings.SplitN(kv, "=", 2)
		if parts[0] != "" {
			p[parts[0]] = parts[1]
		}
	}
	return
}

type PeerHandler func(context.Context, Peer)

func SkipSelf(h PeerHandler) PeerHandler {
	return func(ctx context.Context, p Peer) {
		if !p.IsSelf {
			h(ctx, p)
		}
	}
}

type Controller struct {
	instance *Instance

	port uint

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

func NewController(instance *Instance, port uint) (*Controller, error) {
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		return nil, err
	}
	return &Controller{
		instance: instance,
		port:     port,
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
	server, err := zeroconf.Register(c.instance.Name, PhotoscopeSVCName, "local.", int(c.port), propertiesAsTXT(c.instance.Properties), nil)
	if err != nil {
		logger.Error("Failed to publish zeroconf service: %s", zap.Error(err))
	}
	defer server.Shutdown()

	logger.Info("Looking for peers...")
	for {
		peerCh := make(chan *zeroconf.ServiceEntry)
		browseCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		if err := c.resolver.Browse(browseCtx, PhotoscopeSVCName, "", peerCh); err != nil {
			logger.Error("Failed to browse mDNS services", zap.Error(err))
			return
		}
		logger.Info("mDNS browsing", zap.String("service", PhotoscopeSVCName))

		for {
			select {
			case p := <-peerCh:
				if p == nil {
					continue
				}
				c.peerDiscovered(ctx, p)
			case <-browseCtx.Done():
				logger.Info("mDNS browser terminated")
				break
			case <-c.done:
				logger.Info("Shutting down")
				return
			}
		}
	}
}

func (c *Controller) Shutdown() {
	close(c.done)
}

func (c *Controller) peerDiscovered(ctx context.Context, p *zeroconf.ServiceEntry) {
	c.peerLock.Lock()
	defer c.peerLock.Unlock()

	id := findID(p.Text)
	peer := Peer{
		Name:       p.Instance,
		ID:         id,
		Type:       p.Service,
		URL:        asURL(p),
		Properties: propertiesFromTXT(p.Text),
		IsSelf:     c.instance.ID == id,
	}

	if _, found := c.peers[peer.Name]; !found {
		c.peers[peer.Name] = peer
		logging.From(ctx).Info("Peer detected",
			zap.String("peer.instance", peer.Name),
			zap.Stringer("peer.ID", peer.ID),
			zap.String("peer.URL", peer.URL),
			zap.String("peer.type", peer.Type),
			zap.String("peer.hostname", p.HostName),
			zap.Strings("text", p.Text))
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

func findID(txt []string) InstanceID {
	for _, t := range txt {
		kv := strings.SplitN(t, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "id" {
			return InstanceID(kv[1])
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
