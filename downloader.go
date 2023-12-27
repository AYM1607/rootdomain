package rootdomain

import (
	"bytes"
	_ "embed"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Include a pre-downloaded valid list of suffixes for environments where there's
// no internet access. This also speeds up startup.
//
//go:embed public_suffix_list.dat
var defaultSuffixData string

func generateTldTrie(suffixData string) *Trie {
	ts := strings.Split(suffixData, "\n")
	newMap := make(map[string]*Trie)
	rootNode := &Trie{ExceptRule: false, ValidTld: false, matches: newMap}
	for _, t := range ts {
		if t != "" && !strings.HasPrefix(t, "//") {
			t = strings.TrimSpace(t)
			exceptionRule := t[0] == '!'
			if exceptionRule {
				t = t[1:]
			}
			addTldRule(rootNode, strings.Split(t, "."), exceptionRule)
		}
	}
	return rootNode
}

// syncSuffixData keeps the suffix data for a given exctractor up to date.
// It does an initial load followed by daily syncs.
func syncSuffixData(e *TLDExtract) {
	// Sync daily.
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()

	// Listen to os signals.
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)

loop:
	for {
		// Logic is placed before waiting for the ticker so we can do an initial
		// load before waiting 24 hors.
		suffixData, err := downloadSuffixData()
		// The error can be safely ignored, TLDs change very slowly.
		if err == nil {
			rootNode := generateTldTrie(string(suffixData))
			e.m.Lock()
			e.rootNode = rootNode
			e.m.Unlock()
		}

		select {
		case <-t.C:
			continue
		case <-s:
			break loop
		}
	}
}

func downloadSuffixData() ([]byte, error) {
	u := "https://publicsuffix.org/list/public_suffix_list.dat"
	resp, err := http.Get(u)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	lines := strings.Split(string(body), "\n")
	var buffer bytes.Buffer

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "//") {
			buffer.WriteString(line)
			buffer.WriteString("\n")
		}
	}

	return buffer.Bytes(), nil
}
