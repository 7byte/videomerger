package internal

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
)

// execCommand 执行命令
func execCommand(name string, arg ...string) error {
	// 创建一个命令
	cmd := exec.Command(name, arg...)
	//cmd.Dir = workpath
	slog.Debug("执行命令", "命令参数", cmd)
	// 执行命令
	out, err := cmd.CombinedOutput()
	// 打印命令输出
	if string(out) != "" {
		slog.Debug("指令执行完成", "执行结果", string(out))
	}
	if err != nil {
		return err
	}
	return nil
}

// MergedVideo 合并视频文件
func MergedVideo(inputFilesPath, outputPath string) error {
	slog.Info("执行文件合并", "输入文件列表", inputFilesPath)
	// 执行指令 ffmpeg -f concat -i files.txt -c copy !name!.mov
	if err := execCommand("ffmpeg", "-stats", "-f", "concat", "-safe", "0", "-i", inputFilesPath, "-c", "copy", outputPath); err != nil {
		slog.Error("合并文件失败", "输入文件列表", inputFilesPath, "失败原因", err)
		return err
	}
	slog.Info("合并文件成功", "输出文件路径", outputPath)
	return nil
}

type Segment struct {
	Start float64
	End   float64
}

// ExtractSegments 使用 FFmpeg 截取片段
func ExtractSegments(inputPath, outputPath string, segments []Segment) error {
	slog.Info("开始截取片段", "输入文件", inputPath)
	for idx, segment := range segments {
		start, duration := segment.Start, segment.End-segment.Start
		outputFileName := filepath.Join(outputPath, fmt.Sprintf("output_segment_%d.mp4", idx))
		err := execCommand("ffmpeg", "-i", inputPath, "-ss", fmt.Sprintf("%.2f", start), "-t", fmt.Sprintf("%.2f", duration), "-c", "copy", outputFileName)
		if err != nil {
			slog.Error("截取片段失败", "输入文件", inputPath, "失败原因", err)
			return err
		}
		slog.Info("截取片段成功", "输出文件", outputFileName)
	}
	return nil
}