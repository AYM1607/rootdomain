package rootdomain

import (
	"fmt"
	"log"
	"testing"
)

var (
	cache      = "/tmp/tld.cache"
	tldExtract *TLDExtract
	err        error
)

func init() {
	tldExtract, err = New(true)
	if err != nil {
		log.Fatal(err)
	}
}

func assert(url string, expected *Result, returned *Result, t *testing.T) {
	if (expected.flag == returned.flag) && (expected.sld == returned.sld) && (expected.sub == returned.sub) && (expected.tld == returned.tld) {
		return
	}
	t.Errorf("%s;expected:%+v;returned:%+v", url, expected, returned)
}
func aTestA(t *testing.T) {
	result, _ := tldExtract.Extract("9down.cc.html&amp;sa=u&amp;ei=4sfsul-ximsb4ateiicaag&amp;ved=0cbkqfjac&amp;usg=afqjcnfmetjm8-gpgyszv9l1h6_5p369yg/wp-content/themes/airfolio/scripts/timthumb.php")
	fmt.Println(result)
}

func TestAll(t *testing.T) {
	cases := map[string]*Result{"http://www.google.com": &Result{flag: Domain, sub: "www", sld: "google", tld: "com"},
		"https://www.google.com.hk/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&cad=rja&ved=0CDQQFjAA&url=%68%74%74%70%3a%2f%2f%67%72%6f%75%70%73%2e%67%6f%6f%67%6c%65%2e%63%6f%6d%2f%67%72%6f%75%70%2f%67%6f%6c%61%6e%67%2d%6e%75%74%73%2f%62%72%6f%77%73%65%5f%74%68%72%65%61%64%2f%74%68%72%65%61%64%2f%62%31%61%36%65%31%66%38%37%30%32%62%33%31%31%62&ei=okjQULibA9GYiAfk3IDYDw&usg=AFQjCNFVxgAwHXnmEJWVURboSTiygUMTaQ&sig2=3AIxkh4TR5QYWGXCJtBSZg": &Result{flag: Domain, sub: "www", sld: "google", tld: "com.hk"},
		"http://joe.blogspot.co.uk":             &Result{flag: Domain, sub: "", sld: "joe", tld: "blogspot.co.uk"},
		"git+ssh://www.github.com:8443/":        &Result{flag: Domain, sub: "www", sld: "github", tld: "com"},
		"http://www.!github.com:8443/":          &Result{flag: Malformed},
		"http://www.theregister.co.uk":          &Result{flag: Domain, sub: "www", sld: "theregister", tld: "co.uk"},
		"http://media.forums.theregister.co.uk": &Result{flag: Domain, sub: "media.forums", sld: "theregister", tld: "co.uk"},
		"http://216.22.project.coop/":           &Result{flag: Domain, sub: "216.22", sld: "project", tld: "coop"},
		"http://Gmail.org/":                     &Result{flag: Domain, sld: "gmail", tld: "org"},
		"http://wiki.info/":                     &Result{flag: Domain, sld: "wiki", tld: "info"},
		"http://wiki.information/":              &Result{flag: Malformed},
		"http://wiki/":                          &Result{flag: Malformed},
		"http://258.15.32.876":                  &Result{flag: Malformed},
		"http://www.cgs.act.edu.au/":            &Result{flag: Domain, sub: "www", sld: "cgs", tld: "act.edu.au"},
		"http://www.metp.net.cn":                &Result{flag: Domain, sub: "www", sld: "metp", tld: "net.cn"},
		"http://net.cn":                         &Result{flag: Malformed},
		"http://google.com?q=cats":              &Result{flag: Domain, sub: "", sld: "google", tld: "com"},
		"https://mail.google.com/mail":          &Result{flag: Domain, sub: "mail", sld: "google", tld: "com"},
		"ssh://mail.google.com/mail":            &Result{flag: Domain, sub: "mail", sld: "google", tld: "com"},
		"//mail.google.com/mail":                &Result{flag: Domain, sub: "mail", sld: "google", tld: "com"},
		"mail.google.com/mail":                  &Result{flag: Domain, sub: "mail", sld: "google", tld: "com"},
		"9down.cc&amp;sa=u&amp;ei=4sfsul-ximsb4ateiicaag&amp;ved=0cbkqfjac&amp;usg=afqjcnfmetjm8-gpgyszv9l1h6_5p369yg/wp-content/themes/airfolio/scripts/timthumb.php": &Result{flag: Domain, sub: "", sld: "9down", tld: "cc"},
		"cy":                  &Result{flag: Malformed},
		"c.cy":                &Result{flag: Domain, sub: "", sld: "c", tld: "cy"},
		"b.c.cy":              &Result{flag: Domain, sub: "b", sld: "c", tld: "cy"},
		"a.b.c.cy":            &Result{flag: Domain, sub: "a.b", sld: "c", tld: "cy"},
		"b.ide.kyoto.jp":      &Result{flag: Domain, sub: "", sld: "b", tld: "ide.kyoto.jp"},
		"a.b.ide.kyoto.jp":    &Result{flag: Domain, sub: "a", sld: "b", tld: "ide.kyoto.jp"},
		"b.c.kobe.jp":         &Result{flag: Domain, sub: "", sld: "b", tld: "c.kobe.jp"},
		"a.b.c.kobe.jp":       &Result{flag: Domain, sub: "a", sld: "b", tld: "c.kobe.jp"},
		"city.kobe.jp":        &Result{flag: Domain, sub: "", sld: "city", tld: "kobe.jp"},
		"city.a.kobe.jp":      &Result{flag: Domain, sub: "", sld: "city", tld: "a.kobe.jp"},
		"blogspot.co.uk":      &Result{flag: Malformed},
		"blah.blogspot.co.uk": &Result{flag: Domain, sub: "", sld: "blah", tld: "blogspot.co.uk"},
	}
	for url, expected := range cases {
		t.Run(url, func(t *testing.T) {
			returned, err := tldExtract.Extract(url)
			if expected.flag != Malformed || err == nil {
				assert(url, expected, returned, t)
			}
		})
	}
}
