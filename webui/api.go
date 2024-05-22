package webui

import (
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-broadcast/config"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

//go:embed dist/*
//go:embed dist/icons/*
//go:embed dist/assets/*
var content embed.FS

// TextFile 定义了.txt文件的信息
type TextFile struct {
	Filename string `json:"filename"`
}

// BatchFile 定义了.bat文件的信息，包括内容
type BatchFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// NewCombinedMiddleware 创建并返回一个带有依赖的中间件闭包
func CombinedMiddleware(config config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/webui/api") {

			// 处理/api/login的POST请求
			if c.Param("filepath") == "/api/login" && c.Request.Method == http.MethodPost {
				HandleLoginRequest(c, config)
				return
			}
			// 处理/api/check-login-status的GET请求
			if c.Param("filepath") == "/api/check-login-status" && c.Request.Method == http.MethodGet {
				HandleCheckLoginStatusRequest(c)
				return
			}
			// 处理/api/run的GET请求
			if c.Param("filepath") == "/api/run" && c.Request.Method == http.MethodGet {
				handleRunCommand(c)
				return
			}
			// 处理 /api/list-files 路由的请求
			if c.Param("filepath") == "/api/list-files" && c.Request.Method == http.MethodGet {
				handleListFiles(c)
				return
			}
			// 处理 /api/new-save 路由的请求
			if c.Param("filepath") == "/api/new-save" && c.Request.Method == http.MethodPost {
				handleCreateSaveFile(c)
				return
			}

		} else {
			// 否则，处理静态文件请求
			// 如果请求是 "/webui/" ，默认为 "index.html"
			filepathRequested := c.Param("filepath")
			if filepathRequested == "" || filepathRequested == "/" {
				filepathRequested = "index.html"
			}

			// 使用 embed.FS 读取文件内容
			filepathRequested = strings.TrimPrefix(filepathRequested, "/")
			data, err := content.ReadFile("dist/" + filepathRequested)
			if err != nil {
				c.String(http.StatusNotFound, "File not found: %v", err)
				return
			}

			mimeType := getContentType(filepathRequested)

			c.Data(http.StatusOK, mimeType, data)
		}
		// 调用c.Next()以继续处理请求链
		c.Next()
	}
}

func getContentType(path string) string {
	// todo 根据需要增加更多的 MIME 类型
	switch filepath.Ext(path) {
	case ".html":
		return "text/html"
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "text/plain"
	}
}

// HandleLoginRequest处理登录请求
func HandleLoginRequest(c *gin.Context, config config.Config) {
	var json struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkCredentials(json.Username, json.Password, config) {
		// 如果验证成功，设置cookie
		cookieValue, err := GenerateCookie()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate cookie"})
			return
		}

		c.SetCookie("login_cookie", cookieValue, 3600*24, "/", "", false, true)

		c.JSON(http.StatusOK, gin.H{
			"isLoggedIn": true,
			"cookie":     cookieValue,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"isLoggedIn": false,
		})
	}
}

func checkCredentials(username, password string, jsonconfig config.Config) bool {
	serverUsername := jsonconfig.Account
	serverPassword := jsonconfig.Password
	fmt.Printf("有用户正尝试使用 用户名:%v 密码:%v 进行登入\n", username, password)
	fmt.Printf("A user is attempting to log in with Username: %v Password: %v\n", username, password)

	fmt.Printf("请使用默认登入用户[%v] 默认密码[%v] 进行登入,不包含[],遇到问题可到QQ群:196173384 请教\n", serverUsername, serverPassword)
	fmt.Printf("please use default account[%v] default password[%v] to login, not include []\n", serverUsername, serverPassword)
	return username == serverUsername && password == serverPassword
}

// HandleCheckLoginStatusRequest 检查登录状态的处理函数
func HandleCheckLoginStatusRequest(c *gin.Context) {
	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		// 如果cookie不存在，而不是返回BadRequest(400)，我们返回一个OK(200)的响应
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(cookieValue)
	if err != nil {
		switch err {
		case ErrCookieNotFound:
			c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie not found"})
		case ErrCookieExpired:
			c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie has expired"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"isLoggedIn": false, "error": "Internal server error"})
		}
		return
	}

	if isValid {
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Invalid cookie"})
	}
}

// handleRunCommand 处理 /run 路由的请求
func handleRunCommand(c *gin.Context) {

	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(cookieValue)
	if err != nil || !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid cookie"})
		return
	}

	// 从请求中提取参数
	params := c.Request.URL.Query()

	// 构建命令行参数
	args := []string{}
	for key, values := range params {
		if len(values) > 0 {
			for _, val := range values {
				if val != "" { // 确保 val 不为空
					if val == "true" && (key == "g" || key == "f" || key == "r") {
						// 对于布尔型参数，如果值为"true"，只添加参数名
						args = append(args, fmt.Sprintf("-%s", key))
						break // 仅需要添加一次参数名
					} else {
						// 否则添加参数名和参数值
						args = append(args, fmt.Sprintf("-%s", key), val)
					}
				}
			}
		}
	}

	// 在Windows上启动新窗口来运行程序
	cmdLine := fmt.Sprintf("cmd.exe /c start %s %s", os.Args[0], strings.Join(args, " "))
	cmd := exec.Command("cmd.exe", "/c", cmdLine)
	err = cmd.Start()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 响应成功启动的信息
	c.JSON(200, gin.H{"message": "Process started successfully"})
}

// handleCreateSaveFile 处理 /new-save 路由的请求
func handleCreateSaveFile(c *gin.Context) {
	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(cookieValue)
	if err != nil || !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid cookie"})
		return
	}

	// 从请求的body中获取参数b（假设参数以JSON形式发送）
	var requestBody struct {
		Filename string `json:"filename"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	// 检查filename是否提供
	if requestBody.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	// 创建文件名添加 "-save.txt"
	saveFileName := fmt.Sprintf("%s-save.txt", requestBody.Filename)
	executablePath, err := os.Executable()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not determine executable path"})
		return
	}
	dirPath := filepath.Dir(executablePath)

	// 创建文件
	filePath := filepath.Join(dirPath, saveFileName)
	file, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Could not create file: %s", err)})
		return
	}
	defer file.Close()

	// 响应文件创建成功的信息
	c.JSON(http.StatusOK, gin.H{"message": "File created successfully", "filePath": filePath})
}

// handleListFiles 处理 /list-files 路由的请求
func handleListFiles(c *gin.Context) {
	executablePath, err := os.Executable()
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to get executable path"})
		return
	}
	dirPath := filepath.Dir(executablePath)

	var textFiles []TextFile
	var batchFiles []BatchFile

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			switch ext {
			case ".txt":
				textFiles = append(textFiles, TextFile{Filename: info.Name()})
			case ".bat":
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				// Decode content from GBK to UTF-8
				decoder := simplifiedchinese.GBK.NewDecoder()
				reader := transform.NewReader(strings.NewReader(string(content)), decoder)
				decodedContent, err := io.ReadAll(reader)
				if err != nil {
					return err
				}
				// Replace boolean flags with true values
				decodedContentStr := string(decodedContent)
				decodedContentStr = strings.ReplaceAll(decodedContentStr, " -g ", " -g true ")
				decodedContentStr = strings.ReplaceAll(decodedContentStr, " -f ", " -f true ")
				decodedContentStr = strings.ReplaceAll(decodedContentStr, " -r ", " -r true ")

				batchFiles = append(batchFiles, BatchFile{
					Filename: info.Name(),
					Content:  decodedContentStr,
				})
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.JSON(200, gin.H{"textFiles": textFiles, "batchFiles": batchFiles})
}
