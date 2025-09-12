package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kartverket/skipctl/pkg/logging"
	"go.yaml.in/yaml/v4"
)

// ChatGPT port of https://raw.githubusercontent.com/yannh/kubeconform/refs/heads/master/scripts/openapi2jsonschema.py with some extras

// Minimal CRD headers to detect CRD docs quickly.
type objectMeta struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// We only decode the pieces we need; schema itself is left as generic maps.
type crdDoc struct {
	objectMeta `yaml:",inline"`
	Spec       struct {
		Group string `yaml:"group"`
		Names struct {
			Kind   string `yaml:"kind"`
			Plural string `yaml:"plural"`
		} `yaml:"names"`
		Versions []struct {
			Name    string         `yaml:"name"`
			Served  bool           `yaml:"served"`
			Storage bool           `yaml:"storage"`
			Schema  map[string]any `yaml:"schema"`
		} `yaml:"versions"`
	} `yaml:"spec"`
}

type versionInfo struct {
	Name    string
	Served  bool
	Storage bool
	Schema  map[string]any
}

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

var (
	flagOutDir = flag.String("outdir", "./schemas", "output directory for JSON Schemas")
	flagURLs   multiFlag
	flagFiles  multiFlag

	flagOnlyVersion multiFlag // regex, repeatable
	flagSkipVersion multiFlag // regex, repeatable
	flagExcludePre  = flag.Bool("exclude-pre", false, "skip prerelease versions (alpha|beta|rc)")

	flagOnlyKind  multiFlag // regex, repeatable
	flagSkipKind  multiFlag // regex, repeatable
	flagOnlyGroup multiFlag // regex, repeatable
	flagSkipGroup multiFlag // regex, repeatable

	flagServedOnly  = flag.Bool("served-only", false, "include only versions with served=true")
	flagStorageOnly = flag.Bool("storage-only", false, "include only the storage=true version per CRD")
	flagOnePerKind  = flag.Bool("one-per-kind", false, "emit one best version per Kind")

	onlyVerREs, skipVerREs     []*regexp.Regexp
	onlyKindREs, skipKindREs   []*regexp.Regexp
	onlyGroupREs, skipGroupREs []*regexp.Regexp

	invalid = regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)
	log     = logging.ConfigureLogging("text", false)
)

func main() {
	flag.Var(&flagURLs, "url", "URL to a CRD YAML (repeatable)")
	flag.Var(&flagFiles, "file", "local path to a CRD YAML (repeatable)")

	flag.Var(&flagOnlyVersion, "only-version", "regex of version names to include (repeatable)")
	flag.Var(&flagSkipVersion, "skip-version", "regex of version names to exclude (repeatable)")
	flag.Var(&flagOnlyKind, "only-kind", "regex of Kinds to include (repeatable)")
	flag.Var(&flagSkipKind, "skip-kind", "regex of Kinds to exclude (repeatable)")
	flag.Var(&flagOnlyGroup, "only-group", "regex of Groups to include (repeatable)")
	flag.Var(&flagSkipGroup, "skip-group", "regex of Groups to exclude (repeatable)")
	flag.Parse()

	compileAll()

	if len(flagURLs) == 0 && len(flagFiles) == 0 {
		fatalf("no inputs provided; use -url and/or -file")
	}

	if err := os.MkdirAll(*flagOutDir, 0o755); err != nil {
		fatalf("creating outdir %s: %v", *flagOutDir, err)
	}

	// Process URLs then files
	for _, u := range flagURLs {
		if err := handleSource(fetchURL(u), fmt.Sprintf("url %s", u)); err != nil {
			fatalf("%v", err)
		}
	}
	for _, f := range flagFiles {
		if err := handleSource(readFile(f), fmt.Sprintf("file %s", f)); err != nil {
			fatalf("%v", err)
		}
	}
}

func compileAll() {
	compile := func(src []string) []*regexp.Regexp {
		var out []*regexp.Regexp
		for _, s := range src {
			re, err := regexp.Compile(s)
			if err != nil {
				fatalf("invalid regex %q: %v", s, err)
			}
			out = append(out, re)
		}
		return out
	}
	onlyVerREs = compile(flagOnlyVersion)
	skipVerREs = compile(flagSkipVersion)
	onlyKindREs = compile(flagOnlyKind)
	skipKindREs = compile(flagSkipKind)
	onlyGroupREs = compile(flagOnlyGroup)
	skipGroupREs = compile(flagSkipGroup)
}

//nolint:gocognit // should be fixed later
func crdAllowed(group, kind string) bool {
	if len(onlyGroupREs) > 0 {
		ok := false
		for _, re := range onlyGroupREs {
			if re.MatchString(group) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(onlyKindREs) > 0 {
		ok := false
		for _, re := range onlyKindREs {
			if re.MatchString(kind) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	for _, re := range skipGroupREs {
		if re.MatchString(group) {
			return false
		}
	}
	for _, re := range skipKindREs {
		if re.MatchString(kind) {
			return false
		}
	}
	return true
}

func versionAllowed(ver string, served, storage bool) bool {
	if len(onlyVerREs) > 0 {
		ok := false
		for _, re := range onlyVerREs {
			if re.MatchString(ver) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if *flagExcludePre && isPreRelease(ver) {
		return false
	}
	for _, re := range skipVerREs {
		if re.MatchString(ver) {
			return false
		}
	}
	if *flagServedOnly && !served {
		return false
	}
	if *flagStorageOnly && !storage {
		return false
	}
	return true
}

func isPreRelease(ver string) bool {
	v := strings.ToLower(ver)
	return strings.Contains(v, "alpha") || strings.Contains(v, "beta") || strings.Contains(v, "rc")
}

var verRE = regexp.MustCompile(`^v(\d+)(?:[^\d]*(\d+))?$`)

//nolint:mnd // that's how stability is ranked
func stabilityRank(ver string) int {
	v := strings.ToLower(ver)
	switch {
	case strings.Contains(v, "alpha"):
		return 0
	case strings.Contains(v, "beta"), strings.Contains(v, "rc"):
		return 1
	default:
		return 2
	}
}

//nolint:mnd // intentional
func parseNums(ver string) (int, int) {
	var major, tail int
	m := verRE.FindStringSubmatch(strings.ToLower(ver))
	if len(m) >= 2 {
		major, _ = strconv.Atoi(m[1])
	}
	if len(m) >= 3 {
		tail, _ = strconv.Atoi(m[2])
	}
	return major, tail
}

func pickBestVersion(cands []versionInfo) *versionInfo {
	if len(cands) == 0 {
		return nil
	}
	var stor []versionInfo
	for _, c := range cands {
		if c.Storage {
			stor = append(stor, c)
		}
	}
	if len(stor) > 0 {
		cands = stor
	}
	var served []versionInfo
	for _, c := range cands {
		if c.Served {
			served = append(served, c)
		}
	}
	if len(served) > 0 {
		cands = served
	}
	sort.SliceStable(cands, func(i, j int) bool {
		ri, rj := stabilityRank(cands[i].Name), stabilityRank(cands[j].Name)
		if ri != rj {
			return ri > rj
		}
		mi, ti := parseNums(cands[i].Name)
		mj, tj := parseNums(cands[j].Name)
		if mi != mj {
			return mi > mj
		}
		return ti > tj
	})
	return &cands[0]
}

//nolint:gocognit,funlen // should be fixed later
func handleSource(r io.ReadCloser, label string) error {
	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()
	if r == nil {
		return fmt.Errorf("%s: nil reader", label)
	}

	dec := yaml.NewDecoder(r)
	for {
		var raw map[string]any
		if err := dec.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("%s: decode yaml: %w", label, err)
		}
		// Skip empty docs
		if len(raw) == 0 {
			continue
		}

		// Quick header probe
		b, _ := yaml.Marshal(raw)
		var meta objectMeta
		_ = yaml.Unmarshal(b, &meta)
		if meta.Kind != "CustomResourceDefinition" {
			continue
		}

		var d crdDoc
		if err := yaml.Unmarshal(b, &d); err != nil {
			return fmt.Errorf("%s: parse CRD meta: %w", label, err)
		}

		group := d.Spec.Group
		kind := d.Spec.Names.Kind
		plural := d.Spec.Names.Plural

		if !crdAllowed(group, kind) {
			log.Info("skipping CRD by filter", "group", group, "kind", kind)
			continue
		}

		var cands []versionInfo
		for _, v := range d.Spec.Versions {
			version := v.Name

			openapi, sok := v.Schema["openAPIV3Schema"].(map[string]any)
			if !sok || openapi == nil {
				// Legacy nesting: spec.validation.openAPIV3Schema
				if alt, vok := v.Schema["validation"].(map[string]any); vok {
					if o2, saltok := alt["openAPIV3Schema"].(map[string]any); saltok {
						openapi = o2
					}
				}
			}
			if openapi == nil {
				log.Info("no openAPIV3Schema for version, skipping", "group", group, "kind", kind, "version", version)
				continue
			}
			if !versionAllowed(version, v.Served, v.Storage) {
				log.Info("skipping version by filter", "group", group, "kind", kind, "version", version, "served", v.Served, "storage", v.Storage)
				continue
			}

			cands = append(cands, versionInfo{
				Name:    version,
				Served:  v.Served,
				Storage: v.Storage,
				Schema:  openapi,
			})
		}
		if len(cands) == 0 {
			log.Info("no versions matched filters for CRD", "group", group, "kind", kind)
			continue
		}

		emit := cands
		if *flagStorageOnly || *flagOnePerKind {
			if best := pickBestVersion(cands); best != nil {
				emit = []versionInfo{*best}
			}
		}

		for _, e := range emit {
			openapi := e.Schema

			// Carry the GVK extension so validators can resolve schema properly.
			addGVKExtension(openapi, group, e.Name, kind)

			// Optionally add a top-level $schema (draft-07 is common with k8s-style schemas).
			if _, ok := openapi["$schema"]; !ok {
				openapi["$schema"] = "http://json-schema.org/draft-07/schema#"
			}

			data, err := marshalCanonicalJSON(openapi)
			if err != nil {
				return fmt.Errorf("%s: marshal schema: %w", label, err)
			}

			fname := sanitize(fmt.Sprintf("%s_%s_%s.json", group, e.Name, kind))
			if group == "" { // unlikely for CRDs; guard anyway
				fname = sanitize(fmt.Sprintf("%s_%s.json", plural, e.Name))
			}
			path := filepath.Join(*flagOutDir, strings.ToLower(fname))
			//nolint:gosec // the permissions are fine for this use case
			if werr := os.WriteFile(path, data, 0o644); werr != nil {
				return fmt.Errorf("write %s: %w", path, werr)
			}
			log.Info("wrote schema", "path", path, "group", group, "version", e.Name, "kind", kind, "served", e.Served, "storage", e.Storage)
		}
	}
	return nil
}

func fetchURL(u string) io.ReadCloser {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		fatalf("build request: %v", err)
	}
	req.Header.Set("Accept", "application/yaml, text/yaml, text/plain; q=0.9, */*; q=0.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fatalf("GET %s: %v", u, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fatalf("GET %s: unexpected status %s", u, resp.Status)
	}
	if ct := resp.Header.Get("Content-Type"); strings.Contains(ct, "text/html") {
		fatalf("URL %s returned HTML, not YAML", u)
	}
	return resp.Body
}

func readFile(path string) io.ReadCloser {
	f, err := os.Open(path)
	if err != nil {
		fatalf("open %s: %v", path, err)
	}
	return f
}

func addGVKExtension(schema map[string]any, group, version, kind string) {
	ex := map[string]any{"group": group, "version": version, "kind": kind}
	var arr []any
	if v, kok := schema["x-kubernetes-group-version-kind"]; kok {
		if s, ok := v.([]any); ok {
			arr = s
		}
	}
	arr = append(arr, ex)
	schema["x-kubernetes-group-version-kind"] = arr
}

// marshalCanonicalJSON encodes a map with deterministic key ordering to keep diffs clean.
func marshalCanonicalJSON(v any) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	// Reorder maps recursively
	ordered := orderKeys(v)
	if err := enc.Encode(ordered); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}

func orderKeys(v any) any {
	switch t := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		// Encode as map[string]any again, but visit in key order
		out := make(map[string]any, len(t))
		for _, k := range keys {
			out[k] = orderKeys(t[k])
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = orderKeys(t[i])
		}
		return out
	default:
		return v
	}
}

func sanitize(s string) string {
	return invalid.ReplaceAllString(s, "-")
}

func fatalf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	log.Error(msg)
	os.Exit(1)
}
