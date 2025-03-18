package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"meeting-mm/config"

	"github.com/google/uuid"
)

// WhisperService 提供语音识别功能
type WhisperService struct {
	modelPath       string
	useLocalWhisper bool
	tempDir         string
}

// NewWhisperService 创建WhisperService实例
func NewWhisperService(cfg *config.Config) *WhisperService {
	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "meeting-mm-whisper")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("创建临时目录失败: %v", err)
		tempDir = os.TempDir()
	}

	return &WhisperService{
		modelPath:       cfg.WhisperModelPath,
		useLocalWhisper: cfg.UseLocalWhisper,
		tempDir:         tempDir,
	}
}

// TranscribeAudio 将音频文件转录为文本
func (s *WhisperService) TranscribeAudio(audioData []byte) (string, error) {
	if !s.useLocalWhisper {
		return s.transcribeWithAPI(audioData)
	}
	return s.transcribeWithPythonWhisper(audioData)
}

// transcribeWithLocalWhisper 使用本地Whisper.cpp进行转录
func (s *WhisperService) transcribeWithLocalWhisper(audioData []byte) (string, error) {
	// 创建临时音频文件
	audioFile := filepath.Join(s.tempDir, uuid.New().String()+".wav")
	if err := os.WriteFile(audioFile, audioData, 0644); err != nil {
		return "", fmt.Errorf("保存音频文件失败: %w", err)
	}
	defer os.Remove(audioFile)

	// 创建输出文件路径
	outputFile := audioFile + ".txt"
	defer os.Remove(outputFile)

	// 构建命令
	whisperPath := filepath.Join("..", "whisper", "main")
	cmd := exec.Command(whisperPath, "-m", s.modelPath, "-f", audioFile, "-otxt")

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("执行Whisper命令失败: %w, 输出: %s", err, string(output))
	}

	// 读取转录结果
	transcript, err := os.ReadFile(outputFile)
	if err != nil {
		// 如果读取文件失败，尝试从命令输出中提取转录结果
		outputStr := string(output)
		if strings.Contains(outputStr, "[Transcription]") {
			parts := strings.Split(outputStr, "[Transcription]")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
		return "", fmt.Errorf("读取转录结果失败: %w", err)
	}

	return string(transcript), nil
}

// transcribeWithPythonWhisper 使用Python版本的Whisper模型进行转录
func (s *WhisperService) transcribeWithPythonWhisper(audioData []byte) (string, error) {
	// 创建临时音频文件
	audioFile := filepath.Join(s.tempDir, uuid.New().String()+".mp3")
	if err := os.WriteFile(audioFile, audioData, 0644); err != nil {
		return "", fmt.Errorf("保存音频文件失败: %w", err)
	}
	defer os.Remove(audioFile)

	// 创建Python脚本文件
	scriptFile := filepath.Join(s.tempDir, "whisper_transcribe.py")
	scriptContent := `
import sys
import os
import whisper
import torch

# 强制使用CPU，并设置为Float32精度
device = "cpu"
torch.set_default_tensor_type(torch.FloatTensor)

# 加载模型
model = whisper.load_model("base", device=device)

# 加载音频
audio_file = sys.argv[1]
audio = whisper.load_audio(audio_file)
audio = whisper.pad_or_trim(audio)

# 生成梅尔频谱图
mel = whisper.log_mel_spectrogram(audio, n_mels=model.dims.n_mels).to(device)

# 检测语言
_, probs = model.detect_language(mel)
detected_language = max(probs, key=probs.get)
print(f"检测到的语言: {detected_language}", file=sys.stderr)

# 解码音频
options = whisper.DecodingOptions(fp16=False)  # 禁用fp16
result = whisper.decode(model, mel, options)

# 输出转录文本
print(result.text)
`
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		return "", fmt.Errorf("创建Python脚本失败: %w", err)
	}
	defer os.Remove(scriptFile)

	// 明确指定使用Anaconda环境中的Python
	pythonCmd := "/Users/xujiawei/anaconda3/bin/python"

	// 如果指定的Python不存在，尝试使用默认的Python
	if _, err := os.Stat(pythonCmd); os.IsNotExist(err) {
		// 查找可用的Python
		paths := []string{
			"python3",
			"python",
		}
		for _, path := range paths {
			if _, err := exec.LookPath(path); err == nil {
				pythonCmd = path
				break
			}
		}
	}

	// 执行Python脚本
	cmd := exec.Command(pythonCmd, scriptFile, audioFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("执行Python Whisper失败: %w, 输出: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

// transcribeWithAPI 使用API进行转录（备用方案）
func (s *WhisperService) transcribeWithAPI(audioData []byte) (string, error) {
	// 这里可以实现调用OpenAI Whisper API或其他语音识别API的逻辑
	// 由于我们优先使用本地Whisper，这里暂时返回错误
	return "", errors.New("API转录功能尚未实现，请启用本地Whisper")
}

// StreamTranscribe 流式转录音频
func (s *WhisperService) StreamTranscribe(audioStream io.Reader) (chan string, error) {
	// 创建结果通道
	resultChan := make(chan string)

	// 创建临时文件
	tempFile := filepath.Join(s.tempDir, uuid.New().String()+".wav")
	file, err := os.Create(tempFile)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}

	// 启动goroutine处理流式转录
	go func() {
		defer close(resultChan)
		defer os.Remove(tempFile)
		defer file.Close()

		// 将音频流写入临时文件
		_, err := io.Copy(file, audioStream)
		if err != nil {
			resultChan <- fmt.Sprintf("错误: 写入音频数据失败: %v", err)
			return
		}

		// 关闭文件以便读取
		file.Close()

		// 读取文件内容
		audioData, err := os.ReadFile(tempFile)
		if err != nil {
			resultChan <- fmt.Sprintf("错误: 读取音频数据失败: %v", err)
			return
		}

		// 转录音频
		transcript, err := s.TranscribeAudio(audioData)
		if err != nil {
			resultChan <- fmt.Sprintf("错误: 转录失败: %v", err)
			return
		}

		// 发送结果
		resultChan <- transcript
	}()

	return resultChan, nil
}
