package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/clientv3"

	"ivxv.ee/conf"
	"ivxv.ee/log"
	"ivxv.ee/storage/etcd"
)

// protocol prepends the used etcd protocol (https://) to the address.
func protocol(addr string) string {
	return "https://" + addr
}

// bootstrap reports if id is one of the storage services responsible for
// bootstrapping the cluster.
func bootstrap(id string, c *etcd.Conf) bool {
	for _, b := range c.Bootstrap {
		if b == id {
			return true
		}
	}
	return false
}

// walDir returns the directory to use for etcd write-ahead logs. It returns
// the default etcd location, but make it explicit to protect against future
// changes.
func walDir(wd string) string {
	return filepath.Join(wd, "etcd", "member", "wal")
}

// firstBoot reports if this is the first time booting this storage server. It
// uses the same approach as etcd internally, i.e., checks if the write-ahead
// log directory contains any files.
func firstBoot(wd string) (bool, error) {
	f, err := os.Open(walDir(wd))
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, OpenWALDirError{Err: err}
	}
	defer f.Close() // nolint: errcheck, ignore close failure of read-only fd.
	files, err := f.Readdirnames(-1)
	if err != nil {
		return false, ReadWALDirError{Err: err}
	}
	return len(files) == 0, nil
}

type member struct {
	id        uint64
	name      string
	clientURL string
	peerURL   string
}

func hexID(id uint64) string {
	return fmt.Sprintf("%x", id)
}

// members retrieves the current cluster member list.
func members(ctx context.Context, client *clientv3.Client, optime time.Duration) (
	[]member, error) {

	log.Log(ctx, ListingMembers{})
	ctx, cancel := context.WithTimeout(ctx, optime)
	defer cancel()
	list, err := client.MemberList(ctx)
	if err != nil {
		return nil, MemberListError{Err: err}
	}
	log.Log(ctx, MemberList{Members: list.Members})

	members := make([]member, len(list.Members))
	for i, pbm := range list.Members {
		m := member{id: pbm.ID, name: pbm.Name}

		switch len(pbm.ClientURLs) {
		case 0: // Member has not been started yet.
		case 1:
			m.clientURL = pbm.ClientURLs[0]
		default:
			// Multiple client URLs are not supported.
			return nil, UnexpectedMemberClientURLCountError{
				ID:   hexID(m.id),
				Name: m.name,
				URLs: pbm.ClientURLs,
			}
		}

		// Multiple peer URLs are not supported.
		if len(pbm.PeerURLs) != 1 {
			return nil, UnexpectedMemberPeerURLCountError{
				ID:   hexID(m.id),
				Name: m.name,
				URLs: pbm.PeerURLs,
			}
		}
		m.peerURL = pbm.PeerURLs[0]
		members[i] = m
	}
	return members, nil
}

// pruneMembers removes all members from the cluster that are not configured.
func pruneMembers(ctx context.Context, client *clientv3.Client, optime time.Duration,
	conf clientv3.Config, members []member, configured []*conf.Service) error {

next:
	for _, m := range members {
		// Check if the member is still configured.
		for _, service := range configured {
			if m.peerURL == protocol(service.PeerAddress) {
				continue next
			}
		}

		// First ping the member's client URL and ensure that it does
		// not respond. If the member has no client URL, then it has
		// not been started yet and can be freely pruned.
		if len(m.clientURL) > 0 {
			log.Log(ctx, PingingClientURL{URL: m.clientURL})

			// Since this member is no longer configured, its
			// client URL is not among the existing client's
			// endpoints. We must create a new ping client for
			// connecting to this client URL.
			conf.Endpoints = []string{m.clientURL}
			pingc, err := clientv3.New(conf)
			if err != nil {
				log.Debug(ctx, PingClientError{Err: err})
			} else {
				// nolint: errcheck, ignore close errors of ping client.
				defer pingc.Close()

				opctx, cancel := context.WithTimeout(ctx, optime)
				defer cancel()
				status, err := pingc.Status(opctx, m.clientURL)
				if err == nil {
					log.Debug(ctx, PingStatus{Status: status})
					return RemovedMemberStillAliveError{
						ID:        hexID(m.id),
						Name:      m.name,
						ClientURL: m.clientURL,
					}
				}
				log.Debug(ctx, PingError{Err: err})
			}
		}

		// Remove stopped and deconfigured member from cluster.
		log.Log(ctx, PruningMember{ID: hexID(m.id), Name: m.name, URL: m.peerURL})
		opctx, cancel := context.WithTimeout(ctx, optime)
		defer cancel()
		if _, err := client.MemberRemove(opctx, m.id); err != nil {
			return RemoveMemberError{Err: err}
		}
		log.Log(ctx, PrunedMember{ID: hexID(m.id)})
	}
	return nil
}

// addMember adds a member to the cluster if not already there.
func addMember(ctx context.Context, client *clientv3.Client, optime time.Duration,
	members []member, service *conf.Service) error {

	// Check the the member is not already there.
	peerURL := protocol(service.PeerAddress)
	for _, m := range members {
		if m.peerURL == peerURL {
			// When either the member's name or client URL is not
			// empty, then it must match the desired service's.
			if len(m.name) > 0 && m.name != service.ID ||
				len(m.clientURL) > 0 &&
					m.clientURL != protocol(service.Address) {

				return PeerURLAlreadyInUseError{
					ID:   hexID(m.id),
					Name: m.name,
				}
			}
			return nil
		}
	}

	// Add new member to cluster.
	log.Log(ctx, AddingMember{Name: service.ID, URL: peerURL})
	ctx, cancel := context.WithTimeout(ctx, optime)
	defer cancel()
	added, err := client.MemberAdd(ctx, []string{peerURL})
	if err != nil {
		return AddMemberError{Err: err}
	}
	log.Log(ctx, AddedMember{ID: hexID(added.Member.ID)})
	return nil
}

// updateMembers updates the cluster membership using members, pruneMembers and
// addMember.
func updateMembers(ctx context.Context, conf clientv3.Config, optime time.Duration,
	configured []*conf.Service, add *conf.Service) error {

	// If we are adding a member, then remove its endpoint from the client
	// configuration: we cannot connect to it before it is added.
	if add != nil {
		// Do not mangle the existing endpoints: create a new one.
		endpoints := make([]string, 0, len(conf.Endpoints)-1)
		for _, endpoint := range conf.Endpoints {
			if endpoint == protocol(add.Address) {
				continue
			}
			endpoints = append(endpoints, endpoint)
		}
		conf.Endpoints = endpoints
	}

	// Initialize an etcd client to use.
	log.Log(ctx, UpdatingClusterMembership{Endpoints: conf.Endpoints})
	client, err := clientv3.New(conf)
	if err != nil {
		return EtcdClientError{Err: err}
	}
	defer func() {
		// nolint: vetshadow, false positive.
		if err := client.Close(); err != nil {
			// Only log close error, do not return.
			log.Error(ctx, EtcdClientCloseError{Err: err})
		}
	}()

	// Get the current list of members configured in the cluster.
	members, err := members(ctx, client, optime)
	if err != nil {
		return UpdateListMembersError{Err: err}
	}

	// Prune members which are no longer configured: must be done before
	// adding to put the cluster in a healthy state for membership changes.
	if err := pruneMembers(ctx, client, optime, conf, members, configured); err != nil {
		return UpdatePruneMembersError{Err: err}
	}

	if add != nil {
		// Add this instance to the member list if not already there.
		if err := addMember(ctx, client, optime, members, add); err != nil {
			return UpdateAddMemberError{Err: err}
		}
	}
	log.Log(ctx, ClusterUpToDate{})
	return nil
}
