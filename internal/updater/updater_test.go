package updater

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/internal/version"
)

func TestUpdater_SelfUpdate(t *testing.T) {
	var extension string
	if runtime.GOOS == "windows" {
		extension = ".exe"
	}
	exePath := filepath.Join(os.TempDir(), "lefthook")
	for name, tt := range map[string]struct {
		latestRelease string
		assetName     string
		checksums     string
		opts          Options
		asset         []byte
		err           error
	}{
		"asset not found": {
			latestRelease: "v1.0.0",
			assetName:     "lefthook_1.0.0_darwin_arm64",
			opts: Options{
				Yes:     true,
				Force:   false,
				ExePath: exePath,
			},
			err: errNoAsset,
		},
		"no need to update": {
			latestRelease: "v" + version.Version(false),
			assetName:     "lefthook_1.0.0_darwin_arm64",
			opts: Options{
				Yes:     true,
				Force:   false,
				ExePath: exePath,
			},
			err: nil,
		},
		"forced update but asset not found": {
			latestRelease: "v" + version.Version(false),
			assetName:     "lefthook_1.0.0_darwin_arm64",
			opts: Options{
				Yes:     true,
				Force:   true,
				ExePath: exePath,
			},
			err: errNoAsset,
		},
		"invalid hashsum": {
			latestRelease: "v1.0.0",
			assetName:     "lefthook_1.0.0_" + osNames[runtime.GOOS] + "_" + archNames[runtime.GOARCH] + extension,
			opts: Options{
				Yes:     true,
				Force:   true,
				ExePath: exePath,
			},
			asset: []byte{65, 54, 24, 32, 43, 67, 21},
			checksums: `
				67a5740c6c66d986c5708cddd6bd0bc240db29451646fc4c1398b988dcf7cdfe lefthook_1.0.0_MacOS_arm64
				67a5740c6c66d986c5708cddd6bd0bc240db29451646fc4c1398b988dcf7cdfe lefthook_1.0.0_MacOS_x86_64
				67a5740c6c66d986c5708cddd6bd0bc240db29451646fc4c1398b988dcf7cdfe lefthook_1.0.0_Linux_x86_64
				67a5740c6c66d986c5708cddd6bd0bc240db29451646fc4c1398b988dcf7cdfe lefthook_1.0.0_Linux_arm64
				67a5740c6c66d986c5708cddd6bd0bc240db29451646fc4c1398b988dcf7cdfe lefthook_1.0.0_Windows_x86_64.exe
			`,
			err: errInvalidHashsum,
		},
		"success": {
			latestRelease: "v1.0.0",
			assetName:     "lefthook_1.0.0_" + osNames[runtime.GOOS] + "_" + archNames[runtime.GOARCH] + extension,
			opts: Options{
				Yes:     true,
				Force:   true,
				ExePath: exePath,
			},
			asset: []byte{65, 54, 24, 32, 43, 67, 21},
			checksums: `
				0e1c97246ba1bc8bde78355ae986589545d3c69bf1264d2d3c1835ec072006f6 lefthook_1.0.0_MacOS_arm64
				0e1c97246ba1bc8bde78355ae986589545d3c69bf1264d2d3c1835ec072006f6 lefthook_1.0.0_MacOS_x86_64
				0e1c97246ba1bc8bde78355ae986589545d3c69bf1264d2d3c1835ec072006f6 lefthook_1.0.0_Linux_x86_64
				0e1c97246ba1bc8bde78355ae986589545d3c69bf1264d2d3c1835ec072006f6 lefthook_1.0.0_Linux_arm64
				0e1c97246ba1bc8bde78355ae986589545d3c69bf1264d2d3c1835ec072006f6 lefthook_1.0.0_Windows_x86_64.exe
			`,
			err: nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			file, err := os.Create(tt.opts.ExePath)
			assert.NoError(err)
			file.Close()

			checksumServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					n, werr := w.Write([]byte(tt.checksums))
					assert.Equal(n, len(tt.checksums))
					assert.NoError(werr)
				}))
			defer checksumServer.Close()
			assetServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					n, werr := w.Write(tt.asset)
					assert.Equal(n, len(tt.asset))
					assert.NoError(werr)
				}))
			defer assetServer.Close()

			releaseServer := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.NoError(json.NewEncoder(w).Encode(map[string]interface{}{
						"tag_name": tt.latestRelease,
						"assets": []map[string]string{
							{
								"name":                 tt.assetName,
								"browser_download_url": assetServer.URL,
							},
							{
								"name":                 "lefthook_checksums.txt",
								"browser_download_url": checksumServer.URL,
							},
						},
					}))
				}))
			defer releaseServer.Close()

			upd := Updater{
				client:     releaseServer.Client(),
				releaseURL: releaseServer.URL,
			}

			err = upd.SelfUpdate(context.Background(), tt.opts)

			if tt.err != nil {
				if !errors.Is(err, tt.err) {
					t.Error(err)
				}
			} else {
				assert.NoError(err)

				if tt.asset != nil {
					content, err := os.ReadFile(tt.opts.ExePath)
					assert.NoError(err)

					assert.Equal(content, tt.asset)
				}
			}
		})
	}
}
