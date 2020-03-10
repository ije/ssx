package dnsutil

import (
	"hash/fnv"
	"strings"
	"time"

	"ssx/dns/response"

	"github.com/miekg/dns"
)

type CacheItem struct {
	Rcode              int
	AuthenticatedData  bool
	RecursionAvailable bool
	Answer             []dns.RR
	Ns                 []dns.RR
	Extra              []dns.RR

	origTTL uint32
	stored  time.Time
}

func NewCacheItem(m *dns.Msg, now time.Time, ttl time.Duration) *CacheItem {
	i := new(CacheItem)
	i.Rcode = m.Rcode
	i.AuthenticatedData = m.AuthenticatedData
	i.RecursionAvailable = m.RecursionAvailable
	i.Answer = m.Answer
	i.Ns = m.Ns
	i.Extra = make([]dns.RR, len(m.Extra))
	// Don't copy OPT records as these are hop-by-hop.
	j := 0
	for _, e := range m.Extra {
		if e.Header().Rrtype == dns.TypeOPT {
			continue
		}
		i.Extra[j] = e
		j++
	}
	i.Extra = i.Extra[:j]

	i.origTTL = uint32(ttl.Seconds())
	i.stored = now.UTC()

	return i
}

// ToMsg turns i into a message, it tailors the reply to m.
// The Authoritative bit should be set to 0, but some client stub resolver implementations, most notably,
// on some legacy systems(e.g. ubuntu 14.04 with glib version 2.20), low-level glibc function `getaddrinfo`
// useb by Python/Ruby/etc.. will discard answers that do not have this bit set.
// So we're forced to always set this to 1; regardless if the answer came from the cache or not.
// On newer systems(e.g. ubuntu 16.04 with glib version 2.23), this issue is resolved.
// So we may set this bit back to 0 in the future ?
func (i *CacheItem) ToMsg(m *dns.Msg, now time.Time) *dns.Msg {
	m1 := new(dns.Msg)
	m1.SetReply(m)

	// Set this to true as some DNS clients discard the *entire* packet when it's non-authoritative.
	// This is probably not according to spec, but the bit itself is not super useful as this point, so
	// just set it to true.
	m1.Authoritative = true
	m1.AuthenticatedData = i.AuthenticatedData
	m1.RecursionAvailable = i.RecursionAvailable
	m1.Rcode = i.Rcode

	m1.Answer = make([]dns.RR, len(i.Answer))
	m1.Ns = make([]dns.RR, len(i.Ns))
	m1.Extra = make([]dns.RR, len(i.Extra))

	ttl := uint32(i.TTL(now))
	for j, r := range i.Answer {
		m1.Answer[j] = dns.Copy(r)
		m1.Answer[j].Header().Ttl = ttl
	}
	for j, r := range i.Ns {
		m1.Ns[j] = dns.Copy(r)
		m1.Ns[j].Header().Ttl = ttl
	}
	// newCacheItem skips OPT records, so we can just use i.Extra as is.
	for j, r := range i.Extra {
		m1.Extra[j] = dns.Copy(r)
		m1.Extra[j].Header().Ttl = ttl
	}
	return m1
}

func (i *CacheItem) TTL(now time.Time) int {
	ttl := int(i.origTTL) - int(now.UTC().Sub(i.stored).Seconds())
	return ttl
}

// GetCacheKey returns key under which we store the Cacheitem, -1 will be returned if we don't store the message.
// Currently we do not cache Truncated, errors zone transfers or dynamic update messages.
// qname holds the already lowercased qname.
func GetCacheKey(m *dns.Msg, t response.Type, do bool) (bool, uint64) {
	// We don't store truncated responses.
	if m.Truncated {
		return false, 0
	}

	// Nor errors or Meta or Update
	if t == response.OtherError || t == response.Meta || t == response.Update {
		return false, 0
	}

	h := fnv.New64()

	if do {
		h.Write([]byte{'1'})
	} else {
		h.Write([]byte{'0'})
	}

	qname := strings.ToLower(dns.Name(m.Question[0].Name).String())
	qtype := m.Question[0].Qtype
	h.Write([]byte{byte(qtype >> 8)})
	h.Write([]byte{byte(qtype)})
	h.Write([]byte(qname))

	return true, h.Sum64()
}
