package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/gensokyo-broadcast/config"
	"github.com/hoshinonyaruko/gensokyo-broadcast/sys"
	"github.com/hoshinonyaruko/gensokyo-broadcast/txt"
	"github.com/hoshinonyaruko/gensokyo-broadcast/webui"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type CommandLineArgs struct {
	ApiAddress     string
	GroupListFile  string
	MessageContent string
	DelaySeconds   int
	ChanceToSend   int
	Help           bool
	SaveFilePath   string
	FilterChannel  bool
	FriendMode     bool
	Token          string
	RandomList     bool
}

type GroupList struct {
	Data    []Group     `json:"data"`
	Message string      `json:"message"`
	RetCode int         `json:"retcode"`
	Status  string      `json:"status"`
	Echo    interface{} `json:"echo"`
}

type FriendList struct {
	Data    []FriendData `json:"data"`
	Message string       `json:"message"`
	RetCode int          `json:"retcode"`
	Status  string       `json:"status"`
	Echo    interface{}  `json:"echo"`
}

type FriendData struct {
	Nickname string `json:"nickname"`
	Remark   string `json:"remark"`
	UserID   string `json:"user_id"`
}

type Group struct {
	GroupCreateTime int32  `json:"group_create_time"`
	GroupID         int64  `json:"group_id"`
	GroupLevel      int32  `json:"group_level"`
	GroupMemo       string `json:"group_memo"`
	GroupName       string `json:"group_name"`
	MaxMemberCount  int32  `json:"max_member_count"`
	MemberCount     int32  `json:"member_count"`
}

// getExecutableName 返回当前执行文件的名称
func getExecutableName() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	exeName := filepath.Base(exePath)
	return exeName, nil
}

func saveArgsToBatFile(args CommandLineArgs) {
	// 构建.bat文件名
	batFilename := args.SaveFilePath + ".bat"
	if batFilename == ".bat" { // 检查SaveFilePath是否为空
		return // 如果SaveFilePath为空，则不执行任何操作
	}

	// 开始构建命令行字符串
	var cmdLine strings.Builder
	//cmdLine.WriteString("@echo off\n") // 关闭命令回显

	exeName, err := getExecutableName()
	if err != nil {
		fmt.Println("Error getting executable name:", err)
		return
	}

	cmdLine.WriteString(exeName)

	// 构建命令行参数字符串
	if args.ApiAddress != "" {
		cmdLine.WriteString(fmt.Sprintf(" -a %s", args.ApiAddress))
	}
	if args.MessageContent != "" {
		cmdLine.WriteString(fmt.Sprintf(" -w \"%s\"", args.MessageContent))
	}
	if args.GroupListFile != "" {
		cmdLine.WriteString(fmt.Sprintf(" -p %s", args.GroupListFile))
	}
	if args.FilterChannel {
		cmdLine.WriteString(" -g")
	}
	if args.FriendMode {
		cmdLine.WriteString(" -f")
	}
	if args.DelaySeconds > 0 {
		cmdLine.WriteString(fmt.Sprintf(" -d %d", args.DelaySeconds))
	}
	if args.ChanceToSend > 0 {
		cmdLine.WriteString(fmt.Sprintf(" -c %d", args.ChanceToSend))
	}
	if args.SaveFilePath != "" {
		cmdLine.WriteString(fmt.Sprintf(" -s %s", args.SaveFilePath))
	}
	if args.Token != "" {
		cmdLine.WriteString(fmt.Sprintf(" -t %s", args.Token))
	}
	if args.RandomList {
		cmdLine.WriteString(" -r")
	}
	cmdLine.WriteString("\n")

	// 将命令行参数以GBK编码写入到.bat文件中
	file, err := os.Create(batFilename)
	if err != nil {
		log.Printf("Failed to create .bat file '%s': %v\n", batFilename, err)
		return
	}
	defer file.Close()

	writer := transform.NewWriter(file, simplifiedchinese.GBK.NewEncoder())
	_, err = writer.Write([]byte(cmdLine.String()))
	if err != nil {
		log.Printf("Failed to write to .bat file '%s': %v\n", batFilename, err)
	} else {
		log.Printf("Command line arguments saved to '%s'\n", batFilename)
	}
}

// 解析命令行
func parseArgs() CommandLineArgs {
	var args CommandLineArgs
	flag.StringVar(&args.ApiAddress, "a", "", "HTTP API 的地址")
	flag.StringVar(&args.GroupListFile, "p", "", "群列表的文件名")
	flag.StringVar(&args.MessageContent, "w", "", "要发送的信息")
	flag.IntVar(&args.DelaySeconds, "d", 10, "每条信息推送时间的间隔（秒）")
	flag.IntVar(&args.ChanceToSend, "c", 100, "每个群推送的概率（%百分比）")
	flag.BoolVar(&args.Help, "h", false, "显示帮助信息")
	flag.StringVar(&args.SaveFilePath, "s", "", "读取-save文件路径")
	flag.BoolVar(&args.FilterChannel, "g", false, "gensokyo过滤子频道")
	flag.BoolVar(&args.FriendMode, "f", false, "私聊模式")
	flag.StringVar(&args.Token, "t", "", "access_token")
	flag.BoolVar(&args.RandomList, "r", false, "打乱群/好友列表顺序")
	flag.Parse()

	// 保存命令行参数到.bat文件
	saveArgsToBatFile(args)

	return args
}

// 主函数
func main() {
	if len(os.Args) == 1 {
		// 读取或创建配置
		jsonconfig := config.ReadConfig()

		//cookie数据库
		webui.InitializeDB()

		//给程序整个标题
		sys.SetTitle(jsonconfig.Title + " 作者 早苗狐 答疑群:196173384")

		// 没有命令行参数，启动Web UI
		startWebServer(jsonconfig)

		// 设置信号捕获
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// 等待信号
		<-sigChan
		// 可以执行退出程序
		// 正常退出程序
		os.Exit(0)
	} else {
		// 有命令行参数，执行原有逻辑
		runCommandLineLogic()
	}
}

func startWebServer(jsonconfig config.Config) {

	r := gin.Default()

	//webui和它的api
	webuiGroup := r.Group("/webui")
	{
		webuiGroup.GET("/*filepath", webui.CombinedMiddleware(jsonconfig))
		webuiGroup.POST("/*filepath", webui.CombinedMiddleware(jsonconfig))
		webuiGroup.PUT("/*filepath", webui.CombinedMiddleware(jsonconfig))
		webuiGroup.DELETE("/*filepath", webui.CombinedMiddleware(jsonconfig))
		webuiGroup.PATCH("/*filepath", webui.CombinedMiddleware(jsonconfig))
	}

	// 创建一个http.Server实例(主服务器)
	httpServer := &http.Server{
		Addr:    "0.0.0.0:" + jsonconfig.Port,
		Handler: r,
	}

	if jsonconfig.UseHttps {
		fmt.Printf("webui-api运行在 HTTPS 端口 %v\n", jsonconfig.Port)
		// 在一个新的goroutine中启动主服务器
		go func() {
			// 定义默认的证书和密钥文件名 自签名证书
			certFile := "cert.pem"
			keyFile := "key.pem"
			if jsonconfig.Cert != "" && jsonconfig.Key != "" {
				certFile = jsonconfig.Cert
				keyFile = jsonconfig.Key
			}
			// 使用 HTTPS
			if err := httpServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}

		}()
	} else {
		fmt.Printf("webui-api运行在 HTTP 端口 %v\n", jsonconfig.Port)
		// 在一个新的goroutine中启动主服务器
		go func() {
			// 使用HTTP
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}()
	}
	fmt.Printf("快捷访问:http://127.0.0.1:" + jsonconfig.Port + "/webui")
}

func runCommandLineLogic() {
	ts := txt.GetInstance()
	args := parseArgs()
	if args.Help {
		showHelp()
		return
	}
	executeTaskBasedOnArgs(ts, args)
}

func showHelp() {
	fmt.Println("命令行参数说明：")
	fmt.Println("-a  HTTP API 的地址。示例: -a http://localhost:8080")
	fmt.Println("-p  指定群列表的txt文件名(不包括.txt后缀)。示例: -p group_list")
	fmt.Println("-w  要发送的信息内容。如果包含.txt则尝试从对应的txt文件中读取内容。示例: -w message.txt 或 -w '这是一条消息'||'这是另一条消息'")
	fmt.Println("-s  必须,存档名,指定-save文件路径,用于断点续发。示例: -s 本次任务代号,指定新文件代表从头开始任务。不需要加-save和后缀。")
	fmt.Println("-d  *每条信息推送时间间隔（秒）。示例: -d 15, 默认为10秒。")
	fmt.Println("-c  *每个群推送的概率（百分比）。示例: -c 50, 默认为100%，即总是推送。")
	fmt.Println("-h  *显示帮助信息。不需要值，仅标志存在即可。")
	fmt.Println("-g  *QQ开放平台频道智能选择,ture=每个频道首个文字子频道广播,false=全部子频道都发送广播,不需要值，仅标志存在即可。")
	fmt.Println("-f  *私聊模式,仅限发送通知,不要发送骚扰信息。请遵守调用限制.")
	fmt.Println("-t  *access_token,如果你设置了http的密钥则需要这个参数.")
	fmt.Println("-r  *打乱群和好友列表的顺序.")
}

func executeTaskBasedOnArgs(ts *txt.TxtStore, args CommandLineArgs) {
	// 根据参数执行逻辑
	var groupIDs []int64
	var err error
	var filename string
	// 根据提供的参数执行不同的逻辑
	if args.GroupListFile == "" {
		// 从API获取群列表并保存
		if !args.FriendMode {
			groupIDs, filename, err = fetchAndSaveGroupList(args.ApiAddress, args.SaveFilePath, args.FilterChannel, args.Token, args.RandomList)
			if err != nil {
				log.Fatalf("Failed to read group list from file: %v", err)
			}
		} else {
			groupIDs, filename, err = fetchAndSaveFriendList(args.ApiAddress, args.SaveFilePath, args.Token, args.RandomList)
			if err != nil {
				log.Fatalf("Failed to read group list from file: %v", err)
			}
		}
	} else if args.GroupListFile != "" {
		// 从文件读取群列表
		groupIDs, err = readGroupListFromTS(ts, args.GroupListFile, args.RandomList)
		if err != nil {
			log.Fatalf("Failed to read group list from file: %v", err)
		}
		// 输出从文件读取到的群号数量
		fmt.Printf("从文件%s读取了群列表,%d个群或好友\n", args.GroupListFile, len(groupIDs))
	}
	// 处理消息内容
	message, err := handleMessageContent(ts, args.MessageContent)
	if err != nil {
		log.Fatalf("Error handling message content: %v", err)
	}
	// 发送消息并更新保存文件
	err = sendMessageAndUpdateSaveFile(ts, filename, args.ApiAddress, groupIDs, message, args.DelaySeconds, args.ChanceToSend, args.SaveFilePath, args.FriendMode, args.Token)
	if err != nil {
		log.Fatalf("Error sending messages: %v", err)
	}
}

func sendGroupMessage(apiURL string, groupID int64, userID int64, message string, token string) (string, error) {
	// 首先替换\n和%0A为占位符
	placeholder := "\xFF\xFE"
	message = strings.Replace(message, "\n", placeholder, -1)
	message = strings.Replace(message, "\\n", placeholder, -1)
	message = strings.Replace(message, "%0A", placeholder, -1)

	// 然后将字符串转换为字节切片
	byteMessage := []byte(message)

	// 替换占位符为真正的换行符字节序列（CRLF）
	crlf := []byte{13, 10}
	byteMessage = bytes.ReplaceAll(byteMessage, []byte(placeholder), crlf)

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"group_id": groupID,
		"message":  string(byteMessage),
		"user_id":  userID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	baseurl := apiURL + "/send_group_msg"
	if token != "" {
		baseurl += "?access_token=" + token
	}
	// 发送POST请求
	resp, err := http.Post(baseurl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	responseContent := string(responseBody)

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return responseContent, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	return responseContent, nil
}

func sendPrivateMessage(apiURL string, userID int64, message string, token string) (string, error) {
	// 首先替换\n和%0A为占位符
	placeholder := "\xFF\xFE"
	message = strings.Replace(message, "\n", placeholder, -1)
	message = strings.Replace(message, "\\n", placeholder, -1)
	message = strings.Replace(message, "%0A", placeholder, -1)

	// 然后将字符串转换为字节切片
	byteMessage := []byte(message)

	// 替换占位符为真正的换行符字节序列（CRLF）
	crlf := []byte{13, 10}
	byteMessage = bytes.ReplaceAll(byteMessage, []byte(placeholder), crlf)

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"message": string(byteMessage),
		"user_id": userID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}
	baseurl := apiURL + "/send_private_msg"
	if token != "" {
		baseurl += "?access_token=" + token
	}
	// 发送POST请求
	resp, err := http.Post(baseurl, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	responseContent := string(responseBody)

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return responseContent, fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	return responseContent, nil
}

func parseAndPossiblyRandomize(body []byte, randomlist bool) (*GroupList, error) {
	var groupList GroupList
	if err := json.Unmarshal(body, &groupList); err != nil {
		log.Printf("Failed to unmarshal JSON: %v", err)
		return nil, err
	}

	// 如果 randomlist 为 true，则打乱 Data 数组
	if randomlist {
		rand.Shuffle(len(groupList.Data), func(i, j int) {
			groupList.Data[i], groupList.Data[j] = groupList.Data[j], groupList.Data[i]
		})
	}

	return &groupList, nil
}

func parseAndPossiblyRandomizeFriends(body []byte, randomlist bool) (*FriendList, error) {
	var groupList FriendList
	if err := json.Unmarshal(body, &groupList); err != nil {
		log.Printf("Failed to unmarshal JSON: %v", err)
		return nil, err
	}

	// 如果 randomlist 为 true，则打乱 Data 数组
	if randomlist {
		rand.Shuffle(len(groupList.Data), func(i, j int) {
			groupList.Data[i], groupList.Data[j] = groupList.Data[j], groupList.Data[i]
		})
	}

	return &groupList, nil
}

// 定义从HTTP API获取群列表并保存的函数，返回群列表和可能的错误
func fetchAndSaveGroupList(apiURL string, SaveFilePath string, isgensokyo bool, token string, randomlist bool) ([]int64, string, error) {
	// 构建获取群列表的URL
	url := apiURL + "/get_group_list"
	if token != "" {
		url += "?access_token=" + token
	}

	// 发送HTTP GET请求
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch group list: %v", err)
		return nil, "", err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, "", err
	}

	// 解析JSON到结构体
	groupList, err := parseAndPossiblyRandomize(body, randomlist)
	if err != nil {
		log.Println("Error processing JSON:", err)
	} else {
		log.Printf("Processed group list: %+v", groupList)
	}

	// 创建文件以保存群列表
	filename := fmt.Sprintf("%d-%s.txt", time.Now().Unix(), SaveFilePath)
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return nil, "", err
	}
	defer file.Close()

	// 准备收集群ID
	var groupIDs []int64
	if !isgensokyo {
		// 写入群ID到文件并收集群ID
		for _, group := range groupList.Data {
			_, err := file.WriteString(strconv.FormatInt(group.GroupID, 10) + "\n")
			if err != nil {
				log.Printf("Failed to write to file: %v", err)
				return nil, "", err
			}
			groupIDs = append(groupIDs, group.GroupID)
		}
	} else {
		// 特殊逻辑
		lookingForSubChannel := false // 用于标记是否正在寻找首个子频道

		for _, group := range groupList.Data {
			// 检查GroupName是否为空，如果为空，直接加入
			if group.GroupName == "" {
				groupIDs = append(groupIDs, group.GroupID)
				//log.Printf("GroupName为空，已添加GroupID：%d", group.GroupID)
				lookingForSubChannel = false // 重置标记
				_, err := file.WriteString(strconv.FormatInt(group.GroupID, 10) + "\n")
				if err != nil {
					log.Printf("Failed to write to file: %v", err)
					return nil, "", err
				}
				continue
			}

			if strings.HasPrefix(group.GroupName, "*") {
				// 如果GroupName以*开头，开始寻找以&开头的首个子频道
				log.Printf("检测到频道，GroupID: %d, 频道名称: %s", group.GroupID, group.GroupName)
				lookingForSubChannel = true // 设置标记为true
			} else if lookingForSubChannel {
				// 仅当我们正在寻找首个子频道，并且GroupName以&开头时，才处理
				if strings.HasPrefix(group.GroupName, "&") {
					groupIDs = append(groupIDs, group.GroupID)
					log.Printf("检测到首个子频道GroupID: %d, 子频道名称: %s", group.GroupID, group.GroupName)
					lookingForSubChannel = false // 找到后重置标记
					_, err := file.WriteString(strconv.FormatInt(group.GroupID, 10) + "\n")
					if err != nil {
						log.Printf("Failed to write to file: %v", err)
						return nil, "", err
					}
				}
			}
			// 如果不是以上任一情况，则继续循环
		}
	}

	log.Printf("Group list saved to %s\n", filename)

	return groupIDs, filename, nil // 返回群ID数组和nil表示没有错误
}

// 定义从HTTP API获取好友列表并保存的函数，返回群列表和可能的错误
func fetchAndSaveFriendList(apiURL string, SaveFilePath string, token string, randomlist bool) ([]int64, string, error) {
	// 构建获取群列表的URL
	url := apiURL + "/get_friend_list"
	if token != "" {
		url += "?access_token=" + token
	}

	// 发送HTTP GET请求
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch group list: %v", err)
		return nil, "", err
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, "", err
	}

	// 解析JSON到结构体
	groupList, err := parseAndPossiblyRandomizeFriends(body, randomlist)
	if err != nil {
		log.Println("Error processing JSON:", err)
	} else {
		log.Printf("Processed group list: %+v", groupList)
	}

	// 创建文件以保存群列表
	filename := fmt.Sprintf("%d-%s.txt", time.Now().Unix(), SaveFilePath)
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return nil, "", err
	}
	defer file.Close()

	// 准备收集群ID
	var FriendIDs []int64

	// 写入群ID到文件并收集群ID
	for _, friend := range groupList.Data {
		_, err := file.WriteString(friend.UserID + "\n")
		if err != nil {
			log.Printf("Failed to write to file: %v", err)
			return nil, "", err
		}
		friendid64, _ := strconv.ParseInt(friend.UserID, 10, 64)
		FriendIDs = append(FriendIDs, friendid64)
	}

	log.Printf("Friends list saved to %s\n", filename)

	return FriendIDs, filename, nil // 返回群ID数组和nil表示没有错误
}

// ts是txt单例对象，且GetFileContent方法返回一个包含文件每行内容的字符串数组和一个错误
// readGroupListFromTS 从文本存储中读取群列表，并根据 randomlist 决定是否随机打乱
func readGroupListFromTS(ts *txt.TxtStore, filename string, randomlist bool) ([]int64, error) {
	// 从 ts 单例获取文件内容
	lines, err := ts.GetFileContent(filename)
	if err != nil {
		log.Println("Error:", err)
		return nil, err
	}

	// 解析字符串数组内容为群ID列表
	var groupIDs []int64
	for _, line := range lines {
		if line == "" {
			continue
		}
		groupID, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			log.Printf("Invalid group ID in content: %v", err)
			continue
		}
		groupIDs = append(groupIDs, groupID)
	}

	// 如果 randomlist 为 true，则打乱 groupIDs 列表
	if randomlist {
		rand.Shuffle(len(groupIDs), func(i, j int) {
			groupIDs[i], groupIDs[j] = groupIDs[j], groupIDs[i]
		})
	}

	return groupIDs, nil
}

// handleMessageContent 处理消息内容，如果是.txt文件，则从对应的txt文件中读取
func handleMessageContent(ts *txt.TxtStore, content string) ([]string, error) {
	// 检查content是否以".txt"后缀结尾
	if strings.HasSuffix(content, ".txt") {
		// 移除".txt"后缀获取实际的文件名
		filenameWithoutExtension := strings.TrimSuffix(content, ".txt")
		// 使用修改后的文件名调用GetFileContent方法
		lines, err := ts.GetFileContent(filenameWithoutExtension)
		if err != nil {
			return nil, err
		}
		// 输出从文件读取到的行数信息
		fmt.Printf("从文件'%s.txt'读取了%d行自定义回复\n", filenameWithoutExtension, len(lines))
		return lines, nil
	} else {
		// 如果content不是文件名，则根据'||'分割消息内容
		messages := strings.Split(content, "||")
		// 处理分割结果，去除两端的空格
		for i, msg := range messages {
			messages[i] = strings.TrimSpace(msg)
		}
		return messages, nil
	}
}

func sendMessageAndUpdateSaveFile(ts *txt.TxtStore, filename string, apiURL string, groupIDs []int64, messages []string, delay int, chance int, saveFile string, isfriend bool, token string) error {
	progressFilename := saveFile
	fmt.Printf("执行发送任务,目标%d个群或好友\n", len(groupIDs))
	for _, groupID := range groupIDs {
		// 检查是否已有发送记录
		sent, err := hasSendRecord(ts, progressFilename, groupID)
		if err != nil {
			log.Printf("发送记录未创建，可能是第一次本任务。")
		}
		if sent {
			log.Printf("Message to group %d already sent, skipping\n", groupID)
			continue
		}

		// 随机选择一个消息发送
		message := messages[rand.Intn(len(messages))]

		var sendResult string
		// 根据概率决定是否发送
		if rand.Intn(100) < chance {
			if !isfriend {
				// 调用API发送消息
				sendResult, err = sendGroupMessage(apiURL, groupID, 0, message, token) // UserID设置为0
				if err != nil {
					log.Printf("Failed to send message to group %d: %v\n", groupID, err)
					sendResult = "失败: " + err.Error() // 记录失败状态
				}
				// 在发送后输出目标群和消息内容
				fmt.Printf("正在向群号为%d的群发送消息: %s\n", groupID, message)
			} else {
				// 调用API发送消息
				sendResult, err = sendPrivateMessage(apiURL, groupID, message, token) // 这里的groupID是UserID
				if err != nil {
					log.Printf("Failed to send message to friends %d: %v\n", groupID, err)
					sendResult = "失败: " + err.Error() // 记录失败状态
				}
				// 在发送后输出目标群和消息内容
				fmt.Printf("正在向ID号为%d的用户发送私聊消息: %s\n", groupID, message)
			}
			fmt.Printf("发送状态: %s\n", sendResult)

			// 记录到保存文件
			appendSaveFile(filename, progressFilename, groupID, sendResult, time.Now().Format("2006-01-02 15:04:05"))
		} else {
			log.Printf("Skipped sending message to group %d due to chance setting\n", groupID)
		}

		// 延迟发送下一条消息
		time.Sleep(time.Duration(delay) * time.Second)
	}

	return nil
}

// copyIfNeeded 检查进度文件是否存在，如果不存在则从原始群号列表复制
func copyIfNeeded(originalFilename, progressFilename string) error {
	if _, err := os.Stat(progressFilename); os.IsNotExist(err) {
		originalFile, err := os.Open(originalFilename)
		if err != nil {
			return fmt.Errorf("failed to open original file: %v", err)
		}
		defer originalFile.Close()

		progressFile, err := os.Create(progressFilename)
		if err != nil {
			return fmt.Errorf("failed to create progress file: %v", err)
		}
		defer progressFile.Close()

		scanner := bufio.NewScanner(originalFile)
		for scanner.Scan() {
			_, err := progressFile.WriteString(scanner.Text() + "\n")
			if err != nil {
				return fmt.Errorf("failed to write to progress file: %v", err)
			}
		}
	} else if err != nil {
		return fmt.Errorf("failed to check progress file: %v", err)
	}

	return nil
}

// appendSaveFile 用于更新进度文件
func appendSaveFile(originalFilename, baseFilename string, groupID int64, sendResult, timestamp string) {
	progressFilename := baseFilename + "-save.txt"
	// 确保进度文件存在，如果不存在则从原始文件复制
	err := copyIfNeeded(originalFilename, progressFilename)
	if err != nil {
		log.Printf("Error preparing progress file: %v", err)
		return
	}
	// 读取进度文件，更新匹配的群号行
	lines, err := readLines(progressFilename)
	if err != nil {
		log.Printf("Failed to read progress file '%s': %v", progressFilename, err)
		return
	}
	// 更新文件内容
	groupIDStr := strconv.FormatInt(groupID, 10)
	updated := false
	for i, line := range lines {
		if strings.HasPrefix(line, groupIDStr) {
			// 去除sendResult中除了末尾以外的所有换行符
			cleanSendResult := strings.ReplaceAll(sendResult, "\n", "")
			lines[i] = fmt.Sprintf("%s %s %s", line, cleanSendResult, timestamp)
			updated = true
			break
		}
	}
	// 如果找到并更新了群号，重写进度文件
	if updated {
		err = writeLines(lines, progressFilename)
		if err != nil {
			log.Printf("Failed to write updated content to progress file '%s': %v", progressFilename, err)
		}
	}
}

// readLines 从给定的文件名读取所有行
func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
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

// writeLines 将字符串切片写入给定的文件名，覆盖其内容
func writeLines(lines []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

// hasSendRecord 检查给定群号是否已经有发送记录
func hasSendRecord(ts *txt.TxtStore, baseFilename string, groupID int64) (bool, error) {
	// 从TxtStore获取文件内容
	lines, err := ts.GetFileContent(baseFilename + "-save")
	if err != nil {
		return false, fmt.Errorf("failed to get content from TxtStore for '%s': %v", baseFilename, err)
	}

	groupIDStr := strconv.FormatInt(groupID, 10)
	// 正则表达式匹配 YYYY-MM-DD 格式的日期
	dateRegex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

	for _, line := range lines {
		//fmt.Printf("test:%v", line)
		if strings.HasPrefix(line, groupIDStr) {
			// 检查这一行是否包含日期格式的字符串，即是否包含发送时间戳
			if dateRegex.MatchString(line) {
				return true, nil
			}
			break
		}
	}
	return false, nil
}
