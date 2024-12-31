package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/7byte/videomerger/internal"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
)

type mergeFLags struct {
	cronSpec  string
	dateRange string
}

var defaultMergeFlags mergeFLags

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge video files",
	Long:  "Merge video files",
	Run: func(cmd *cobra.Command, args []string) {
		if defaultMergeFlags.cronSpec != "" {
			// 定时任务
			slog.Info("定时任务启动", "cron spec", defaultMergeFlags.cronSpec)
			c := cron.New(cron.WithChain(
				cron.Recover(cron.DefaultLogger),
				cron.SkipIfStillRunning(cron.DefaultLogger),
			))
			c.AddFunc(defaultMergeFlags.cronSpec, func() {
				slog.Info("定时任务执行")
				merge()
			})
			c.Start()
			select {}
		} else {
			merge()
		}
	},
}

func init() {
	mergeCmd.Flags().StringVarP(&defaultMergeFlags.cronSpec, "cron_spec", "c", "", "cron表达式，默认为空，即只运行一次，cron语法参考：https://en.m.wikipedia.org/wiki/Cron")
	mergeCmd.Flags().StringVarP(&defaultMergeFlags.dateRange, "date_range", "r", "", "日期范围，如\"20060102-20060202\"，开始日期和结束日期都可以为空，默认为空，即合并所有文件")
	rootCmd.AddCommand(mergeCmd)
}

func merge() {
	if inputPath == "" {
		slog.Error("input path is empty")
		return
	}

	// 检查输出目录是否存在，目录不存在时创建目录
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err := os.MkdirAll(outputPath, os.ModePerm)
		if err != nil {
			slog.Error("Failed to create directory", "error", err)
			return
		}
		slog.Info("Directory created:", "path", outputPath)
	} else if err != nil {
		slog.Error("Error checking directory", "error", err)
		return
	}

	// 解析日期范围
	var start, end string
	if defaultMergeFlags.dateRange != "" {
		start, end = parseDateRange(defaultMergeFlags.dateRange)
		if start == "" && end == "" {
			slog.Error("Invalid date range", "date range", defaultMergeFlags.dateRange)
			return
		}
	}

	// 找到所有视频文件
	filelist, err := findAllVedio(inputPath, start, end)
	if err != nil {
		return
	}
	if len(filelist) == 0 {
		slog.Error("No video files found")
		return
	}

	// 合并视频文件
	err = mergeVideoFiles(outputPath, filelist)
	if err != nil {
		return
	}
}

var validDate = regexp.MustCompile(`^\d{8}$`)

// 解析日期范围
func parseDateRange(dateRange string) (string, string) {
	if dateRange == "" {
		return "", ""
	}
	i := strings.Index(dateRange, "-")
	if i == -1 {
		return "", ""
	}
	start, end := dateRange[:i], dateRange[i+1:]
	if (start != "" && !validDate.MatchString(start)) || (end != "" && !validDate.MatchString(end)) {
		return "", ""
	}
	return start, end
}

// 小米监控视频文件名格式，如：20240706212250_20240706213600.mp4
var videofileReg = regexp.MustCompile(`^\d{14}_\d{14}\.mp4$`)

// 遍历目录，找到所有视频文件
func findAllVedio(inputPath, startdate, enddate string) (map[string][]string, error) {
	// 保存文件列表用于合并，key为文件日期，value为文件路径列表
	filelist := make(map[string][]string)
	nowdate := time.Now().Format("20060102")
	// 遍历目录
	// TODO：记录合并进度，支持增量合并
	err := filepath.WalkDir(inputPath, func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if !videofileReg.MatchString(d.Name()) {
			return nil
		}
		// 合并有效路径
		date := d.Name()[:8] // 获取文件日期
		// 只合并当天之前的文件
		if date >= nowdate {
			return nil
		}
		if startdate != "" && date < startdate {
			return nil
		}
		if enddate != "" && date > enddate {
			return nil
		}
		slog.Info("遍历目录", "发现有效视频文件", d.Name())
		filelist[date] = append(filelist[date], "file '"+path+"'\n")
		return nil
	})
	if err != nil {
		slog.Error("遍历目录失败：", "失败原因", err)
		return nil, err
	}
	return filelist, nil
}

func mergeVideoFiles(outputPath string, filelist map[string][]string) error {
	merge := func(date string, files []string) error {
		fileName := date + ".mp4"
		outputFilePath := filepath.Join(outputPath, fileName)
		// 检查是否已经合并
		if _, err := os.Stat(outputFilePath); err == nil {
			slog.Info("文件已存在", "文件路径", outputFilePath)
			return nil
		}
		// 按日期排序
		sort.StringSlice(files).Sort()
		content := strings.Join(files, "")
		// 生成文件列表
		inputFilesPath := filepath.Join(os.TempDir(), fmt.Sprintf("files_%s.txt", date))
		if err := saveFileList(inputFilesPath, content); err != nil {
			slog.Error("保存文件列表失败", "失败原因", err)
			return err
		}
		err := internal.MergedVideo(inputFilesPath, outputFilePath)
		return err
	}
	for date, files := range filelist {
		if err := merge(date, files); err != nil {
			return err
		}
	}
	return nil
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
