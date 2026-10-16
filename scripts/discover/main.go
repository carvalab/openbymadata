// Command discover compares what the open.bymadata.com.ar web app calls
// with what this library implements, and reports what to add, remove or
// fix. It prints a Markdown report and exits 1 when there is something to
// act on.
//
//	go run ./scripts/discover          # report only
//	go run ./scripts/discover -probe   # also call each new endpoint live
//
// The web app is BYMA's reference client: its JavaScript bundle lists every
// endpoint, the "Options" header each one needs, and the request body it
// sends. Endpoints we decided not to implement live in skipped.txt.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	site    = "https://open.bymadata.com.ar"
	freeAPI = site + "/vanoms-be-core/rest/api/bymadata/free/"
)

var (
	reMainJS      = regexp.MustCompile(`src="(main\.[a-f0-9]+\.js)"`)
	reInterceptor = regexp.MustCompile(`if\(\[([^\]]*)\]\.some\(\w+=>\w+\.url\.includes\(\w+\)\)\)\{\w+\.Options="([a-z-]+)"`)
	reFreePath    = regexp.MustCompile(`"/?bymadata/free/([^"?]+)`)
	reBody        = regexp.MustCompile(`"/bymadata/free/([^"?]+)",(\{[^{}]*\})`)
	reConst       = regexp.MustCompile(`(path\w+)\s*=\s*"([^"]+)"`)
	reSection     = regexp.MustCompile(`(path\w+):\s*"([a-z-]+)"`)
	reQuoted      = regexp.MustCompile(`"([^"]*)"`)
)

type endpoint struct {
	path    string
	section string // "Options" header value; "" if the interceptor doesn't list it
	body    string // request body literal from the bundle, if found
}

func main() {
	probe := flag.Bool("probe", false, "call each new endpoint against the live API")
	recheck := flag.Bool("recheck", false, "with -probe, also call the skipped endpoints")
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	bundleName, bundle, err := fetchBundle()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch web app:", err)
		os.Exit(2)
	}
	web := parseBundle(bundle)
	lib, libSections, err := parseLibrary(filepath.Join(*root, "internal/api/endpoints.go"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	skipped, err := parseSkipped(filepath.Join(*root, "scripts/discover/skipped.txt"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	var added, mismatched []string
	var newEndpoints []endpoint
	for _, ep := range web {
		libPath, ok := matchLibrary(ep.path, lib)
		if !ok {
			if _, skip := skipped[ep.path]; !skip {
				newEndpoints = append(newEndpoints, ep)
			}
			continue
		}
		if ep.section != "" && libSections[libPath] != ep.section {
			mismatched = append(mismatched, fmt.Sprintf("`%s`: web app sends `%s`, library sends `%s`", libPath, ep.section, orNone(libSections[libPath])))
		}
	}
	for path := range skipped {
		if !webHas(web, path) {
			fmt.Fprintf(os.Stderr, "note: skipped.txt lists %q, which the web app no longer calls; remove it\n", path)
		}
	}
	var removed []string
	for _, libPath := range sortedKeys(lib) {
		if !webHas(web, libPath) {
			removed = append(removed, libPath)
		}
	}
	for _, ep := range newEndpoints {
		line := fmt.Sprintf("`%s` (Options: `%s`)", ep.path, orNone(ep.section))
		if ep.body != "" {
			line += fmt.Sprintf(" body `%s`", ep.body)
		}
		if *probe {
			line += " → " + probeEndpoint(ep)
		}
		added = append(added, line)
	}

	var rechecked []string
	if *probe && *recheck {
		for _, ep := range web {
			if reason, ok := skipped[ep.path]; ok {
				rechecked = append(rechecked, fmt.Sprintf("`%s` (%s) → %s", ep.path, reason, probeEndpoint(ep)))
			}
		}
	}

	fmt.Printf("# BYMA web app vs openbymadata\n\nBundle: `%s`, %d API paths, %d implemented, %d skipped.\n", bundleName, len(web), len(lib), len(skipped))
	section("New endpoints (implement, or add to scripts/discover/skipped.txt with a reason)", added)
	section("Endpoints the web app no longer calls (check if BYMA removed them)", removed)
	section("Options header mismatches (fix sectionHeader in internal/api/endpoints.go)", mismatched)

	if *recheck {
		section("Skipped endpoints, called again (move any that now return data out of skipped.txt)", rechecked)
	}

	if len(added)+len(removed)+len(mismatched) > 0 {
		os.Exit(1)
	}
	fmt.Println("\nNothing to do.")
}

func fetchBundle() (string, string, error) {
	index, err := get(site + "/")
	if err != nil {
		return "", "", err
	}
	m := reMainJS.FindStringSubmatch(index)
	if m == nil {
		return "", "", fmt.Errorf("main.*.js not found in index.html; the web app layout changed")
	}
	bundle, err := get(site + "/" + m[1])
	return m[1], bundle, err
}

func get(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: HTTP %d", url, resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// parseBundle collects every free-API path in the bundle with the Options
// header from the request interceptor and the request body, when present.
func parseBundle(js string) []endpoint {
	eps := map[string]*endpoint{}
	add := func(path string) *endpoint {
		path = normalize(path)
		if eps[path] == nil {
			eps[path] = &endpoint{path: path}
		}
		return eps[path]
	}
	for _, m := range reInterceptor.FindAllStringSubmatch(js, -1) {
		for _, q := range reQuoted.FindAllStringSubmatch(m[1], -1) {
			add(q[1]).section = m[2]
		}
	}
	for _, m := range reFreePath.FindAllStringSubmatch(js, -1) {
		add(m[1])
	}
	for _, m := range reBody.FindAllStringSubmatch(js, -1) {
		add(m[1]).body = m[2]
	}
	// The interceptor matches by substring, so a fuller path from the
	// body/free-path scan inherits the section of the fragment it contains.
	for _, ep := range eps {
		if ep.section != "" {
			continue
		}
		for _, other := range eps {
			if other.section != "" && (strings.HasSuffix(ep.path, "/"+other.path) || strings.HasPrefix(other.path, ep.path+"/") || strings.HasSuffix(other.path, "/"+ep.path)) {
				ep.section = other.section
				break
			}
		}
	}
	out := make([]endpoint, 0, len(eps))
	for _, ep := range eps {
		if ep.path != "" {
			out = append(out, *ep)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].path < out[j].path })
	return out
}

func normalize(p string) string {
	p = strings.TrimPrefix(p, "/")
	if i := strings.IndexAny(p, "?"); i >= 0 {
		p = p[:i]
	}
	return strings.TrimSuffix(p, "/")
}

func parseLibrary(file string) (map[string]bool, map[string]string, error) {
	src, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, err
	}
	names := map[string]string{}
	lib := map[string]bool{}
	for _, m := range reConst.FindAllStringSubmatch(string(src), -1) {
		names[m[1]] = m[2]
		lib[m[2]] = true
	}
	sections := map[string]string{}
	for _, m := range reSection.FindAllStringSubmatch(string(src), -1) {
		sections[names[m[1]]] = m[2]
	}
	return lib, sections, nil
}

// parseSkipped reads "path  # reason" lines.
func parseSkipped(file string) (map[string]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		path, reason, _ := strings.Cut(line, "#")
		out[normalize(strings.TrimSpace(path))] = strings.TrimSpace(reason)
	}
	return out, sc.Err()
}

// matchLibrary reports the library path that serves web path w. The web app
// often references a fragment ("historical-series/history") or a parent
// ("chart/historical-series") of the full path the library uses.
func matchLibrary(w string, lib map[string]bool) (string, bool) {
	for l := range lib {
		if pathsMatch(w, l) {
			return l, true
		}
	}
	return "", false
}

func pathsMatch(w, l string) bool {
	return w == l || strings.HasSuffix(l, "/"+w) || strings.HasPrefix(l, w+"/")
}

func webHas(web []endpoint, l string) bool {
	for _, ep := range web {
		if pathsMatch(ep.path, l) {
			return true
		}
	}
	return false
}

// probeEndpoint calls the endpoint with the web app's body, filling
// variables with sample values, and summarizes the answer. Endpoints that
// reject POST with 405 are retried as GET with sample query parameters,
// which is how the chart endpoints work.
func probeEndpoint(ep endpoint) string {
	result := call(ep, http.MethodPost, freeAPI+ep.path, sampleBody(ep.body))
	if result != "HTTP 405" {
		return result
	}
	now := time.Now()
	query := fmt.Sprintf("?symbol=GGAL&query=GGAL&limit=5&resolution=D&from=%d&to=%d", now.AddDate(0, 0, -30).Unix(), now.Unix())
	return "POST 405, GET " + call(ep, http.MethodGet, freeAPI+ep.path+query, nil)
}

func call(ep endpoint, method, url string, body []byte) string {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", site)
	req.Header.Set("Referer", site+"/")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	if ep.section != "" {
		req.Header.Set("Options", ep.section)
	}
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return "200, " + shape(raw)
}

// sampleBody turns a JS object literal such as {symbol:this.symbol,page_number:1}
// into JSON, using sample values for variables.
func sampleBody(literal string) []byte {
	samples := map[string]any{"symbol": "GGAL", "settlementType": "2", "currency": "ARS", "date": time.Now().Format("2006-01-02"), "value": "GGAL"}
	obj := map[string]any{"Content-Type": "application/json"}
	for _, part := range strings.Split(strings.Trim(literal, "{}"), ",") {
		k, v, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		k = strings.Trim(strings.TrimSpace(k), `"`)
		v = strings.TrimSpace(v)
		var lit any
		switch {
		case json.Unmarshal([]byte(v), &lit) == nil:
			obj[k] = lit
		case samples[k] != nil:
			obj[k] = samples[k]
		default:
			obj[k] = ""
		}
	}
	b, _ := json.Marshal(obj)
	return b
}

func shape(raw []byte) string {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return fmt.Sprintf("non-JSON (%d bytes)", len(raw))
	}
	if m, ok := v.(map[string]any); ok {
		if data, ok := m["data"]; ok {
			v = data
		}
	}
	switch x := v.(type) {
	case []any:
		if len(x) == 0 {
			return "empty list"
		}
		if first, ok := x[0].(map[string]any); ok {
			return fmt.Sprintf("list[%d], fields: %s", len(x), strings.Join(sortedKeys(first), ", "))
		}
		return fmt.Sprintf("list[%d]", len(x))
	case map[string]any:
		return "object, fields: " + strings.Join(sortedKeys(x), ", ")
	}
	return fmt.Sprintf("%v", v)
}

func section(title string, lines []string) {
	fmt.Printf("\n## %s\n\n", title)
	if len(lines) == 0 {
		fmt.Println("None.")
		return
	}
	for _, l := range lines {
		fmt.Println("- " + l)
	}
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
