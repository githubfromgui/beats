package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirkon/goproxy/gomod"
)

type vendorPackage struct {
	Package []vendorEntry `json:"package"`
}

type vendorEntry struct {
	CheckSumSHA1 string `json:"checksumSHA1"`
	Path         string `json:"path"`
	Revision     string `json:"revision"`
	RevisionTime string `json:"revisionTime"`
	Version      string `json:"version"`
	VersionExact string `json:"versionExact"`
}

var (
	beatsPath        = filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "elastic", "beats")
	modulesCachePath = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	vendorPath       = filepath.Join(beatsPath, "vendor")
	goModPath        = filepath.Join(beatsPath, "go.mod")
)

func main() {
	ret := 0

	fmt.Println("Diffing files under vendor and go modules cache")
	fmt.Println("Go modules cache location:", modulesCachePath)
	fmt.Println("Vendor location:", vendorPath)

	m, err := goModulesFile()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	vendorFile, err := vendorJSON()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filepath.Walk(vendorPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			pathInCache := strings.TrimPrefix(path, vendorPath+"/")
			if _, ok := m.Require[pathInCache]; !ok {
				return nil
			}

			required := pathInCache + "@" + m.Require[pathInCache]
			if m.Replace[pathInCache] != nil {
				dep, _ := m.Replace[pathInCache].(gomod.Dependency)
				required = dep.Path + "@" + dep.Version
			}

			filepath.Walk(path, func(vendoredPath string, vendorInfo os.FileInfo, vendorErr error) error {
				if vendorInfo.IsDir() {
					return nil
				}

				underCache := strings.TrimPrefix(vendoredPath, vendorPath+"/"+pathInCache)
				fullPath := filepath.Join(modulesCachePath, required, underCache)

				vendoredContents, err := openFileForDiffing(vendoredPath)
				if err != nil {
					fmt.Println(err)
					return err
				}
				modulesContents, err := openFileForDiffing(fullPath)
				if err != nil {
					fmt.Println(err)
					return err
				}

				res := bytes.Compare(vendoredContents, modulesContents)
				if res != 0 {
					fmt.Println("MODULE", pathInCache)
					fmt.Println("Vendored:", vendoredPath)
					fmt.Println("go modules:", fullPath)
					fmt.Println("different files")
					for _, entry := range vendorFile.Package {
						if strings.HasPrefix(entry.Path, pathInCache) {
							revTime := strings.ReplaceAll(entry.RevisionTime, "-", "")
							revTime = strings.ReplaceAll(revTime, "Z", "")
							revTime = strings.ReplaceAll(revTime, "T", "")
							revTime = strings.ReplaceAll(revTime, ":", "")
							rev := entry.Revision[:12]
							fmt.Printf("%s-%s\n", revTime, rev)
							break
						}
					}
					fmt.Println("---------")
					ret = 1
					return nil
				}

				return nil
			})
		}

		return nil
	})

	os.Exit(ret)
}

func goModulesFile() (*gomod.Module, error) {
	c, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}
	return gomod.Parse(goModPath, c)
}

func vendorJSON() (*vendorPackage, error) {
	d, err := ioutil.ReadFile(vendorPath + "/vendor.json")
	if err != nil {
		return nil, err
	}

	var vendorFile vendorPackage
	err = json.Unmarshal(d, &vendorFile)
	if err != nil {
		return nil, err
	}
	return &vendorFile, nil
}

func openFileForDiffing(path string) ([]byte, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while reading from file", path, err)
	}
	return contents, nil
}
