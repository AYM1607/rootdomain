package rootdomain

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// used for Result.Flag
const (
	Malformed = iota
	Domain
)

type Result struct {
	Flag int
	Sub  string
	Sld  string
	Tld  string
}

type TLDExtract struct {
	rootNode   *Trie
	debug      bool
	noValidate bool // do not validate URL schema
	noStrip    bool // do not strip .html suffix from URL
	m          sync.RWMutex
}

var (
	schemaregex = regexp.MustCompile(`^([abcdefghijklmnopqrstuvwxyz0123456789\+\-\.]+:)?//`)
	domainregex = regexp.MustCompile(`^[a-z0-9-\p{L}]{1,63}$`)
)

// New creates a new *TLDExtract, it may be shared between goroutines, we usually need a single instance in an application.
func New(debug bool) (*TLDExtract, error) {
	rootNode := generateTldTrie(defaultSuffixData)
	extractor := &TLDExtract{rootNode: rootNode, debug: debug}
	go syncSuffixData(extractor)
	return extractor, nil
}

func (extract *TLDExtract) Extract(u string) *Result {
	extract.m.RLock()
	defer extract.m.RUnlock()

	input := u
	u = strings.ToLower(u)
	// TODO: Since this is meant to be used in an SNI context this filtering
	// can probably be done in linear time instead of with a regex.
	u = schemaregex.ReplaceAllString(u, "")
	i := strings.Index(u, "@")
	if i != -1 {
		u = u[i+1:]
	}
	index := strings.IndexFunc(u, func(r rune) bool {
		switch r {
		case '&', '/', '?', ':', '#':
			return true
		}
		return false
	})
	if index != -1 {
		u = u[0:index]
	}
	if extract.debug {
		fmt.Printf("%s;%s\n", u, input)
	}
	return extract.extract(u)
}

func (extract *TLDExtract) extract(url string) *Result {
	domain, tld := extract.extractTld(url)
	if tld == "" {
		return &Result{Flag: Malformed}
	}
	sub, root := subdomain(domain)
	if domainregex.MatchString(root) {
		return &Result{Flag: Domain, Sld: root, Sub: sub, Tld: tld}
	}
	return &Result{Flag: Malformed}
}

func (extract *TLDExtract) extractTld(url string) (domain, tld string) {
	spl := strings.Split(url, ".")
	tldIndex, validTld := extract.getTldIndex(spl)
	if validTld {
		domain = strings.Join(spl[:tldIndex], ".")
		tld = strings.Join(spl[tldIndex:], ".")
	} else {
		domain = url
	}
	return
}

func (extract *TLDExtract) getTldIndex(labels []string) (int, bool) {
	t := extract.rootNode
	parentValid := false
	for i := len(labels) - 1; i >= 0; i-- {
		lab := labels[i]
		n, found := t.matches[lab]
		_, starfound := t.matches["*"]

		switch {
		case found && !n.ExceptRule:
			parentValid = n.ValidTld
			t = n
		// Found an exception rule
		case found:
			fallthrough
		case parentValid:
			return i + 1, true
		case starfound:
			parentValid = true
		default:
			return -1, false
		}
	}
	return -1, false
}

// return sub domain,root domain
func subdomain(d string) (string, string) {
	ps := strings.Split(d, ".")
	l := len(ps)
	if l == 1 {
		return "", d
	}
	return strings.Join(ps[0:l-1], "."), ps[l-1]
}
