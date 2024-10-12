package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	inputPath  string
	outputPath string
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge video files",
	Long:  "Merge video files",
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	mergeCmd.Flags().StringVarP(&inputPath, "input_path", "i", "", "input path of video files")
	mergeCmd.Flags().StringVarP(&outputPath, "output_path", "o", ".", "output path of merged video, default is current path")
	rootCmd.AddCommand(mergeCmd)
}

// 小米监控视频文件名格式，如：20240706212250_20240706213600.mp4
var videofileReg = regexp.MustCompile(`\d{14}_\d{14}\.mp4`)

func run() {
	if inputPath == "" {
		slog.Error("input path is empty")
		return
	}

	// 检查输出目录是否存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		// 目录不存在，创建目录
		err := os.MkdirAll(outputPath, os.ModePerm)
		if err != nil {
			slog.Error("Failed to create directory:", err)
			return
		}
		slog.Info("Directory created:", "path", outputPath)
	} else if err != nil {
		// 其他错误
		slog.Error("Error checking directory:", err)
		return
	}

	// 保存文件列表用于合并，key为文件日期，value为文件路径列表
	filelist := make(map[string][]string)

	// 遍历目录
	// TODO：记录合并进度，支持增量合并
	err := filepath.WalkDir(inputPath, func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if !videofileReg.MatchString(d.Name()) {
			return nil
		}
		slog.Info("遍历目录", "发现文件", d.Name())

		// 合并有效路径
		date := d.Name()[:8] // 获取文件日期
		filelist[date] = append(filelist[date], "file '"+path+"'\n")

		return nil
	})
	if err != nil {
		// 目录遍历失败，终止合并
		slog.Error("遍历目录失败：", "失败原因", err)
		return
	}

	if len(filelist) == 0 {
		slog.Error("No video files found")
		return
	}

	merge := func(date string, files []string) error {
		// 按时间排序
		sort.StringSlice(files).Sort()
		content := strings.Join(files, "")
		// 生成文件列表
		inputFilesPath := filepath.Join(os.TempDir(), fmt.Sprintf("files_%s.txt", date))
		if err := saveFileList(inputFilesPath, content); err != nil {
			slog.Error("保存文件列表失败", "失败原因", err)
			return err
		}
		outputFilePath := filepath.Join(outputPath, date+".mp4")
		err := saveMergedVideo(inputFilesPath, outputFilePath)
		return err
	}

	for date, files := range filelist {
		if err := merge(date, files); err != nil {
			return
		}
	}
}

func saveFileList(filePath, content string) error {
	// 创建一个文件
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入文件内容
	fmt.Fprintln(file, content)

	// 保存文件
	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}

func saveMergedVideo(inputFilesPath, outputFilePath string) error {
	slog.Info("执行文件合并", "输入文件列表", inputFilesPath)
	// 执行指令 ffmpeg -f concat -i files.txt -c copy !name!.mov
	if err := execCommand("ffmpeg", "-f", "concat", "-safe", "0", "-i", inputFilesPath, "-c", "copy", outputFilePath); err != nil {
		slog.Error("合并文件失败", "输入文件列表", inputFilesPath, "失败原因", err)
		return err
	}
	slog.Info("合并文件成功", "输出文件路径", outputFilePath)
	return nil
}

func execCommand(name string, arg ...string) error {
	// 创建一个命令
	cmd := exec.Command(name, arg...)
	//cmd.Dir = workpath
	slog.Debug("执行命令", "命令参数", cmd)
	// 执行命令
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	// 打印命令输出
	if string(out) != "" {
		slog.Debug("指令执行完成", "执行结果", string(out))
	}
	return nil
}
