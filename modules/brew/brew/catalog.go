package brew

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/commandotj/ax-buzz/modules/brew/data"
)

type CatalogSnapshot struct {
	Packages []*data.Package
	CacheAge time.Duration
	Stale    bool
}

var readCatalogCacheAge = catalogCacheAge
var fetchCatalogFormula = fetchFormula
var fetchCatalogCask = fetchCask
var fetchCatalogInstalledFormula = fetchInstalledFormula
var fetchCatalogInstalledCask = fetchInstalledCask

func LoadPackages(ctx context.Context, forceRefresh bool, progress func(string)) (*CatalogSnapshot, error) {
	formulae := make(chan []*apiFormula, 1)
	casks := make(chan []*apiCask, 1)
	formulaInstalls := make(chan []*installInfo, 1)
	caskInstalls := make(chan []*installInfo, 1)
	errors := make(chan error, 2)

	if progress != nil {
		progress("Loading Homebrew formula catalog")
		progress("Loading Homebrew cask catalog")
		progress("Scanning installed formulae and casks")
	}
	go fetchCatalogFormula(ctx, formulae, errors, forceRefresh)
	go fetchCatalogCask(ctx, casks, errors, forceRefresh)
	go fetchCatalogInstalledFormula(false, formulaInstalls)
	go fetchCatalogInstalledCask(false, caskInstalls)

	var formulaData []*apiFormula
	var caskData []*apiCask
	var installedFormulaData []*installInfo
	var installedCaskData []*installInfo
	for completed := 0; completed < 4; {
		select {
		case formulaData = <-formulae:
			completed++
		case caskData = <-casks:
			completed++
		case installedFormulaData = <-formulaInstalls:
			completed++
		case installedCaskData = <-caskInstalls:
			completed++
		case err := <-errors:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	packages := processAllData(formulaData, caskData, apiFormulaAnalytics{}, apiCaskAnalytics{}, installedFormulaData, installedCaskData)
	allBrewPackages = packages
	cacheAge, ok := readCatalogCacheAge()
	if !ok {
		return nil, fmt.Errorf("Homebrew catalog loaded but cache timestamps are unavailable")
	}
	return &CatalogSnapshot{Packages: packages, CacheAge: cacheAge, Stale: cacheAge > urlCacheTtl}, nil
}

func catalogCacheAge() (time.Duration, bool) {
	paths := []string{
		filepath.Join(taproomCacheDir, formulaJwsJson),
		filepath.Join(taproomCacheDir, caskJwsJson),
	}
	var oldest time.Time
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return 0, false
		}
		if oldest.IsZero() || info.ModTime().Before(oldest) {
			oldest = info.ModTime()
		}
	}
	return time.Since(oldest), true
}
