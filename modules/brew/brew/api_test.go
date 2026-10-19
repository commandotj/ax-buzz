package brew

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type brokenBody struct{}

func (brokenBody) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (brokenBody) Close() error             { return nil }
func httpResponse(code int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: code, Status: fmt.Sprintf("%d test", code), Body: body, Header: make(http.Header)}
}
func setTransport(t *testing.T, rt http.RoundTripper) {
	t.Helper()
	old := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: rt}
	t.Cleanup(func() { http.DefaultClient = old })
}

func TestReadCacheAndFetchErrors(t *testing.T) {
	if got := homeCacheDir(func() (string, error) { return "/home/test", nil }); got != filepath.Join("/home/test", ".cache", "taproom") {
		t.Fatalf("home cache=%q", got)
	}
	if got := homeCacheDir(func() (string, error) { return "", errors.New("no home") }); got != ".cache" {
		t.Fatalf("fallback cache=%q", got)
	}
	dir := t.TempDir()
	cache := filepath.Join(dir, "data")
	if readCacheData(cache) != nil {
		t.Fatal("missing cache should be nil")
	}
	if err := os.WriteFile(cache, []byte("cached"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := string(readCacheData(cache)); got != "cached" {
		t.Fatalf("cache=%q", got)
	}
	if readCacheData(dir) != nil {
		t.Fatal("directory read should fail")
	}
	if got, err := fetchUrlWithCache(context.Background(), "https://example.test", cache, false); err != nil || string(got) != "cached" {
		t.Fatalf("cached fetch=%q,%v", got, err)
	}
	setTransport(t, testTransport(func(*http.Request) (*http.Response, error) {
		return httpResponse(200, io.NopCloser(strings.NewReader("body"))), nil
	}))
	if got, err := fetchUrlWithCache(context.Background(), "https://example.test", filepath.Join(dir, "new", "cache"), true); err != nil || string(got) != "body" {
		t.Fatalf("download=%q,%v", got, err)
	}
	parent := filepath.Join(dir, "file")
	_ = os.WriteFile(parent, []byte("x"), 0600)
	if _, err := fetchUrlWithCache(context.Background(), "https://example.test", filepath.Join(parent, "cache"), true); err != nil {
		t.Fatalf("cache write failure should be nonfatal: %v", err)
	}
	if _, err := fetchUrlWithCache(context.Background(), "https://example.test", filepath.Join(parent, "nested", "cache"), true); err != nil {
		t.Fatalf("cache directory failure should be nonfatal: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "cache-as-directory"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := fetchUrlWithCache(context.Background(), "https://example.test", filepath.Join(dir, "cache-as-directory"), true); err != nil {
		t.Fatalf("cache file write failure should be nonfatal: %v", err)
	}
	if _, err := fetchUrlWithCache(context.Background(), ":bad", cache, true); err == nil || !strings.Contains(err.Error(), "create request") {
		t.Fatalf("bad URL err=%v", err)
	}
	setTransport(t, testTransport(func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") }))
	if _, err := fetchUrlWithCache(context.Background(), "https://example.test", cache, true); err == nil || !strings.Contains(err.Error(), "failed to fetch") {
		t.Fatalf("transport err=%v", err)
	}
	setTransport(t, testTransport(func(*http.Request) (*http.Response, error) {
		return httpResponse(503, io.NopCloser(strings.NewReader(""))), nil
	}))
	if _, err := fetchUrlWithCache(context.Background(), "https://example.test", cache, true); err == nil || !strings.Contains(err.Error(), "bad HTTP status") {
		t.Fatalf("status err=%v", err)
	}
	setTransport(t, testTransport(func(*http.Request) (*http.Response, error) { return httpResponse(200, brokenBody{}), nil }))
	if _, err := fetchUrlWithCache(context.Background(), "https://example.test", cache, true); err == nil || !strings.Contains(err.Error(), "failed to read body") {
		t.Fatalf("read err=%v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fetchUrlWithCache(ctx, "https://example.test", cache, true); err == nil {
		t.Fatal("expected canceled request error")
	}
}

func TestFetchJSONTransportErrors(t *testing.T) {
	setTransport(t, testTransport(func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") }))
	values := make(chan map[string]string, 1)
	errs := make(chan error, 1)
	fetchJsonWithCache("https://example.test", filepath.Join(t.TempDir(), "json"), &map[string]string{}, values, errs, true, context.Background())
	if err := <-errs; err == nil {
		t.Fatal("JSON fetch error was lost")
	}
	fetchJwsJsonWithCache("https://example.test", filepath.Join(t.TempDir(), "jws"), &map[string]string{}, make(chan map[string]string, 1), errs, true, context.Background())
	if err := <-errs; err == nil {
		t.Fatal("JWS fetch error was lost")
	}
}

func TestFetchJSONDecodeBranches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "json")
	dataCh := make(chan map[string]string, 1)
	errCh := make(chan error, 1)
	_ = os.WriteFile(path, []byte(`{"a":"b"}`), 0600)
	fetchJsonWithCache("url", path, &map[string]string{}, dataCh, errCh, false, context.Background())
	if (<-dataCh)["a"] != "b" {
		t.Fatal("json data not delivered")
	}
	_ = os.WriteFile(path, []byte("{"), 0600)
	fetchJsonWithCache("url", path, &map[string]string{}, dataCh, errCh, false, context.Background())
	if err := <-errCh; err == nil {
		t.Fatal("expected JSON error")
	}
	_ = os.WriteFile(path, []byte(`{"payload":"{\"a\":\"b\"}"}`), 0600)
	fetchJwsJsonWithCache("url", path, &map[string]string{}, dataCh, errCh, false, context.Background())
	if (<-dataCh)["a"] != "b" {
		t.Fatal("JWS data not delivered")
	}
	_ = os.WriteFile(path, []byte("{"), 0600)
	fetchJwsJsonWithCache("url", path, &map[string]string{}, dataCh, errCh, false, context.Background())
	if err := <-errCh; err == nil {
		t.Fatal("expected JWS envelope error")
	}
	_ = os.WriteFile(path, []byte(`{"payload":"{"}`), 0600)
	fetchJwsJsonWithCache("url", path, &map[string]string{}, dataCh, errCh, false, context.Background())
	if err := <-errCh; err == nil {
		t.Fatal("expected JWS payload error")
	}
}

func TestPublicCatalogFetchWrappers(t *testing.T) {
	oldDir, oldFlag := taproomCacheDir, *flagInvalidateCache
	taproomCacheDir = t.TempDir()
	*flagInvalidateCache = false
	t.Cleanup(func() { taproomCacheDir = oldDir; *flagInvalidateCache = oldFlag })
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(taproomCacheDir, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(formulaJwsJson, `{"payload":"[]"}`)
	write(caskJwsJson, `{"payload":"[]"}`)
	write(formulaAnalyticsJson, `{"items":[]}`)
	write(caskAnalyticsJson, `{"items":[]}`)
	f, fe := make(chan []*apiFormula, 1), make(chan error, 1)
	fetchFormula(context.Background(), f, fe, false)
	if got := <-f; len(got) != 0 {
		t.Fatalf("formula=%v", got)
	}
	c, ce := make(chan []*apiCask, 1), make(chan error, 1)
	fetchCask(context.Background(), c, ce, false)
	if got := <-c; len(got) != 0 {
		t.Fatalf("cask=%v", got)
	}
	fa, fae := make(chan apiFormulaAnalytics, 1), make(chan error, 1)
	fetchFormulaAnalytics(fa, fae)
	if len((<-fa).Items) != 0 {
		t.Fatal("formula analytics not empty")
	}
	ca, cae := make(chan apiCaskAnalytics, 1), make(chan error, 1)
	fetchCaskAnalytics(ca, cae)
	if len((<-ca).Items) != 0 {
		t.Fatal("cask analytics not empty")
	}
	write(formulaJwsJson, "{")
	bad, errs := make(chan []*apiFormula, 1), make(chan error, 1)
	fetchFormula(context.Background(), bad, errs, false)
	if err := <-errs; err == nil {
		t.Fatal("expected wrapper decode error")
	}
}
