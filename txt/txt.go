package txt

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// TxtStore 存储每个.txt文件的内容
type TxtStore struct {
	files map[string][]string
	mu    sync.RWMutex
}

var (
	instance *TxtStore
	once     sync.Once
)

// GetInstance 以单例模式返回TxtStore实例
func GetInstance() *TxtStore {
	once.Do(func() {
		instance = &TxtStore{
			files: make(map[string][]string),
		}
		instance.loadTxtFiles(".")
		go instance.watchTxtFiles(".")
	})
	return instance
}

// loadTxtFiles 加载目录下所有.txt文件
func (ts *TxtStore) loadTxtFiles(dirPath string) {
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error accessing path:", path, "Error:", err)
			return err
		}
		if filepath.Ext(info.Name()) == ".txt" {
			fileContent, err := ts.readFileContent(path)
			if err != nil {
				log.Println("Error reading file:", path, "Error:", err)
				return err
			}
			ts.mu.Lock()
			ts.files[info.Name()] = fileContent
			ts.mu.Unlock()
		}
		return nil
	})
}

// readFileContent 读取文件内容到字符串数组
func (ts *TxtStore) readFileContent(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// watchTxtFiles 监听.txt文件的变更
func (ts *TxtStore) watchTxtFiles(dirPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("Detected file change:", event.Name)
					if filepath.Ext(event.Name) == ".txt" {
						fileContent, err := ts.readFileContent(event.Name)
						if err != nil {
							log.Println("Error reading file:", event.Name, "Error:", err)
							continue
						}
						ts.mu.Lock()
						ts.files[filepath.Base(event.Name)] = fileContent
						ts.mu.Unlock()
					}
				}
			case err := <-watcher.Errors:
				log.Println("Watcher error:", err)
			}
		}
	}()

	err = watcher.Add(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

// GetFileContent 根据文件名返回文件内容
func (ts *TxtStore) GetFileContent(fileName string) ([]string, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	content, ok := ts.files[fileName+".txt"]
	if !ok {
		// 文件不存在时的处理方式
		return nil, fmt.Errorf("file not found: %s", fileName+".txt")
	}
	return content, nil
}
