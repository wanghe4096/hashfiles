package HashFiles

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"github.com/wanghe4096/HashFiles/glob"
	"fmt"
	"crypto/sha1"

)

const (
	MaxSize = 1024
)

type HandleFunc func(os.FileInfo)

var (
	wg    sync.WaitGroup
	mutex sync.RWMutex
	DefaultWriter = os.Stdout
)

type FileInfo struct {
	Path string
	Size int64
}

func genHashCode(data []byte) string {
	return fmt.Sprintf("%x", sha1.Sum(data))
}

// Scans a directory recursively filtering out files that match the fileMatch regexp
func ScanDirectoryForLogfiles(directoryPath string, ignoreFileMatch *glob.Glob) []FileInfo {
	files := make([]FileInfo, 0)
	filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if ignoreFileMatch.MatchString(path) {

			return nil
		}
		files = append(files, FileInfo{Path:path, Size: info.Size()})
		return nil
	})
	return files
}

func appendLog(file FileInfo) {
	mutex.Lock()
	defer mutex.Unlock()
	f, err := os.Open(file.Path)
	defer f.Close()
	if err != nil {
		log.Println(err)
		return
	}

	data := make([]byte, MaxSize)
	code := genHashCode(data)
	fmt.Fprintf(DefaultWriter, "%s, %d, %v \n", f.Name(), file.Size, code)
	wg.Done()
}

func SetOutput(fd *os.File) {
	DefaultWriter = fd
}

func Run(dir string, ignorePattern string) {
	g := glob.Compile(ignorePattern)
	fileList := ScanDirectoryForLogfiles(dir, g)
	for _, v := range fileList {
		wg.Add(1)
		go appendLog(v)
	}
	wg.Wait()
}
