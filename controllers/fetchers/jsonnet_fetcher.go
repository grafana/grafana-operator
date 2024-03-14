package fetchers

import (
	"archive/tar"
	"compress/gzip"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-operator/v5/controllers/config"

	"github.com/google/go-jsonnet"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

// EmbedFSImporter "imports" data from an in-memory embedFS.
type EmbedFSImporter struct {
	Embed embed.FS
	cache map[string]jsonnet.Contents
	mutex sync.Mutex
}

// Import fetches data from an embedFS struct.
func (importer *EmbedFSImporter) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	importer.mutex.Lock()
	defer importer.mutex.Unlock()

	if importer.cache == nil {
		importer.cache = make(map[string]jsonnet.Contents)
	}

	fetchContents := func(getPath, foundAt string) (contents jsonnet.Contents, found string, err error) {
		if content, ok := importer.cache[getPath]; ok {
			return content, getPath, nil
		}

		b, err := importer.Embed.ReadFile(getPath)
		if err != nil {
			return jsonnet.Contents{}, "", err
		}

		file, err := importer.Embed.Open(getPath)
		if err != nil {
			return jsonnet.Contents{}, "", err
		}
		defer func(file fs.File) {
			_ = file.Close()
		}(file)

		importer.cache[foundAt] = jsonnet.MakeContentsRaw(b)

		return jsonnet.MakeContentsRaw(b), foundAt, nil
	}

	var foundContents jsonnet.Contents
	var s string

	findImport := func(root string) error {
		err = fs.WalkDir(importer.Embed, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, importedPath) {
				foundContents, s, err = fetchContents(path, importedPath)
				if err != nil {
					return err
				}
				return filepath.SkipDir
			}
			return nil
		})
		return err
	}

	err = findImport(filepath.Join("grafonnet-lib", "grafonnet"))
	if err != nil || foundContents.Data() != nil {
		return foundContents, s, err
	}

	err = findImport(filepath.Join("grafonnet-lib", "grafonnet-7.0"))
	if err != nil {
		return jsonnet.Contents{}, "", err
	}

	return foundContents, s, nil
}

func FetchJsonnet(dashboard *v1beta1.GrafanaDashboard, envs map[string]string, libsonnet embed.FS) ([]byte, error) {
	if dashboard.Spec.Jsonnet == "" {
		return nil, fmt.Errorf("no jsonnet Content Found, nil or empty string")
	}
	vm := jsonnet.MakeVM()
	for k, v := range envs {
		vm.ExtVar(k, v)
	}

	vm.Importer(&EmbedFSImporter{Embed: libsonnet})

	jsonString, err := vm.EvaluateAnonymousSnippet(dashboard.Name, dashboard.Spec.Jsonnet)
	return []byte(jsonString), err
}

func listDirectoryContents(dirPath string) error {
	fileInfos, err := os.ReadDir(dirPath)
	fmt.Println("Reading directory:", dirPath, "file infos:", fileInfos)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return err
	}
	for _, fileInfo := range fileInfos {
		fmt.Println(fileInfo.Name())
	}

	return nil
}

func generateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	s := base64.URLEncoding.EncodeToString(randomBytes)[:length]
	s = strings.ReplaceAll(s, "-", "a")

	return s, nil
}

func getJsonProjectBuildRoundName(dashboardName string) (string, error) {
	tsNow := strconv.FormatInt(time.Now().Unix(), 10)
	salt, err := generateRandomString(5)
	if err != nil {
		return "", fmt.Errorf("error salt generating as random string: %w", err)
	}
	// Round name generated using 3 parameters: dash name provided from k8s manifest, current timestamp and
	// salt to prevent collisions between simulations calls and/or multiple dashboards with same name
	return fmt.Sprintf("%s-%s-%s", dashboardName, tsNow, salt), nil
}

func getGzipArchiveFileNameWithExtension(fileName string) string {
	return fmt.Sprintf("%s.tar.gz", fileName)
}

func getGzipArchiveFilePath(fileName string) string {
	return filepath.Join(config.GrafanaDashboardsRuntimeBuild, getGzipArchiveFileNameWithExtension(fileName))
}

func getDecompressedGzipArchiveFilePath(fileName string) string {
	return fmt.Sprintf("%s/%s", config.GrafanaDashboardsRuntimeBuild, fileName)
}

func storeByteArrayGzipOnDisk(gzipFileName string, base64EncodedGzipJsonnetProject []byte) (string, error) {
	gzipFileLocalPath := getGzipArchiveFilePath(gzipFileName)

	if err := os.WriteFile(gzipFileLocalPath, base64EncodedGzipJsonnetProject, os.ModePerm); err != nil {
		return "", fmt.Errorf("error writing compressed data to file: %w", err)
	}

	return gzipFileLocalPath, nil
}

func addPrefixToElements(prefix string, array []string) []string {
	result := make([]string, len(array))
	for i, element := range array {
		result[i] = prefix + element
	}
	return result
}

func buildJsonnetProject(buildName string, envs map[string]string, dashboard *v1beta1.GrafanaDashboard) ([]byte, error) {
	if dashboard.Spec.JsonnetProjectBuild == nil {
		return nil, fmt.Errorf("illegal argument: JsonnetProjectBuild is nil")
	}
	if dashboard.Spec.JsonnetProjectBuild.FileName == "" {
		return nil, fmt.Errorf("illegal argument: FileName is empty")
	}
	if dashboard.Spec.JsonnetProjectBuild.GzipJsonnetProject == nil {
		return nil, fmt.Errorf("illegal argument: GzipJsonnetProject is nil")
	}

	jPath := []string{""}
	if dashboard.Spec.JsonnetProjectBuild.JPath != nil {
		jPath = append(jPath, dashboard.Spec.JsonnetProjectBuild.JPath...)
	}

	base64EncodedGzipJsonnetProject := []byte(dashboard.Spec.JsonnetProjectBuild.GzipJsonnetProject)

	gzipFileLocalPath, err := storeByteArrayGzipOnDisk(buildName, base64EncodedGzipJsonnetProject)
	if err != nil {
		fmt.Println("Error placing gzip file on local disk:", err)
		return nil, err
	}

	extractTo := getDecompressedGzipArchiveFilePath(buildName)

	err = untarGzip(gzipFileLocalPath, extractTo)
	if err != nil {
		return nil, fmt.Errorf("error extracting gzip archive: %w", err)
	}

	vm := jsonnet.MakeVM()
	for k, v := range envs {
		vm.ExtVar(k, v)
	}

	jPath = addPrefixToElements(extractTo+"/", jPath)

	vm.Importer(&jsonnet.FileImporter{JPaths: jPath})

	evaluateFilePath := fmt.Sprintf("%s/%s", extractTo, dashboard.Spec.JsonnetProjectBuild.FileName)

	jsonString, err := vm.EvaluateFile(evaluateFilePath)
	if err != nil {
		return nil, fmt.Errorf("error evaluating jsonnet file: %w", err)
	}
	return []byte(jsonString), nil
}

func postJsonnetProjectBuild(buildName string) error {
	buildFolderPath := getDecompressedGzipArchiveFilePath(buildName)
	buildGzipArchivePath := getGzipArchiveFilePath(buildName)

	err := listDirectoryContents(config.GrafanaDashboardsRuntimeBuild)
	if err != nil {
		fmt.Println("Error listing directory contents:", err)
		return err
	}

	deleteFoldersList := []string{buildFolderPath, buildGzipArchivePath}
	fmt.Println("Delete list candidates: ", deleteFoldersList)
	err = deleteFilesAndFolders(deleteFoldersList)
	if err != nil {
		return err
	}

	fmt.Println("Listing dashboards dir after deletion")

	err = listDirectoryContents(config.GrafanaDashboardsRuntimeBuild)
	if err != nil {
		fmt.Println("Error listing directory contents:", err)
		return err
	}
	return nil
}

func BuildProjectAndFetchJsonnetFrom(dashboard *v1beta1.GrafanaDashboard, envs map[string]string) ([]byte, error) {
	jsonnetProjectBuildName, err := getJsonProjectBuildRoundName(dashboard.Name)
	if err != nil {
		return nil, fmt.Errorf("error generating jsonnet project build name: %w", err)
	}

	jsonBytes, err := buildJsonnetProject(jsonnetProjectBuildName, envs, dashboard)
	if postErr := postJsonnetProjectBuild(jsonnetProjectBuildName); postErr != nil {
		fmt.Println("error cleaning up jsonnet project build: %w", postErr)
	}
	return jsonBytes, err
}

// check for path traversal and correct forward slashes
func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) ||
		strings.HasPrefix(p, "/") ||
		strings.Contains(p, "../") {
		return false
	}
	return true
}

func untarGzip(archivePath, extractPath string) error {
	err := os.MkdirAll(extractPath, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return err
	}

	src, err := os.Open(archivePath)
	if err != nil {
		fmt.Println("Error opening archive:", err)
		return err
	}

	// ungzip
	zr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	// untar
	tr := tar.NewReader(zr)

	// uncompress each element
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break // End of archive
		}
		if err != nil {
			return err
		}
		target := header.Name

		if !validRelPath(header.Name) {
			return fmt.Errorf("tar contained invalid name error %q", target)
		}

		target = filepath.Join(extractPath, header.Name)

		// check the type
		switch header.Typeflag {
		// if it's a dir, and it doesn't exist create it (with 0755 permission)
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, os.ModePerm); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer fileToWrite.Close()

			for {
				_, err := io.CopyN(fileToWrite, tr, 4096)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return err
				}
			}
		default:
			fmt.Printf("Unable to untar type : %c in file %s\n", header.Typeflag, target)
		}
	}

	err = unwrapSingleSubdirectory(extractPath)
	if err != nil {
		fmt.Println("Error unwrapping dir", err)
		return err
	}

	return nil
}

func unwrapSingleSubdirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	subDirCount := 0
	fileCount := 0
	var subDirEntry os.DirEntry

	for _, entry := range entries {
		if entry.IsDir() {
			subDirCount++
			subDirEntry = entry
		}
		fileCount++
	}

	if subDirCount != 1 || fileCount > 1 {
		fmt.Println("No unwrapping needed. Either no subdirectories or more than one subdirectory or there are several files in current directory."+
			"dirCount - ", subDirCount, " fileCount - ", fileCount)
		return nil
	}

	subDirPath := filepath.Join(dirPath, subDirEntry.Name())
	if err != nil {
		fmt.Println("Smth goes wrong")
		return err
	}

	err = copyDir(subDirPath, dirPath)
	if err != nil {
		fmt.Println("Error copying dir", err)
		return err
	}

	return os.RemoveAll(subDirPath)
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// Create the destination directory
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err := copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err := copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func deleteFilesAndFolders(paths []string) error {
	for _, path := range paths {
		fmt.Println("Deleting path: ", path)
		err := os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("error during path \"%s\" deletion, error: %w", path, err)
		}
	}
	return nil
}
