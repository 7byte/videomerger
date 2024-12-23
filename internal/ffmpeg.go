package internal

import (
	"fmt"
	"log/slog"
	"os/exec"
)

// ExecCommand 执行命令
func ExecCommand(name string, arg ...string) error {
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

type Segment struct {
	Start float64
	End   float64
}

// ExtractSegments 使用 FFmpeg 截取片段
func ExtractSegments(input string, segments []Segment) error {
	for idx, segment := range segments {
		start, duration := segment.Start, segment.End-segment.Start
		outputFileName := fmt.Sprintf("output_segment_%d.mp4", idx)
		err := ExecCommand("ffmpeg", "-i", input, "-ss", fmt.Sprintf("%.2f", start), "-t", fmt.Sprintf("%.2f", duration), "-c", "copy", outputFileName)
		if err != nil {
			slog.Error("extracting segment", "error", err)
			return err
		}
		slog.Info("segment extracted", "output_file", outputFileName)
	}
	return nil
}
