# 说话人识别（说话人分割）研究

## 概述

说话人识别（Speaker Diarization）是一种音频处理技术，用于回答"谁在什么时候说话"的问题。对于会议纪要系统，这一功能可以帮助我们区分不同发言人的内容，提高会议纪要的可读性和准确性。

## 技术选项

### 1. Whisper模型结合第三方说话人分割库

**方案描述**：
- 使用现有的Whisper模型进行语音转文字
- 使用专门的说话人分割库进行说话人识别
- 将两者的结果合并，生成带有说话人标注的文字转录

**可选库**：
- **PyAnnote**：Facebook/Meta开发的说话人分割库
- **Speechbrain**：提供说话人分割模型的开源语音工具包
- **Resemblyzer**：基于深度学习的说话人识别库

### 2. 集成式解决方案

**方案描述**：
- 使用支持说话人分割的一体化模型
- 直接获取带有说话人信息的转录结果

**可选库**：
- **Whisper-Diarization**：结合了Whisper和说话人分割的项目
- **SpeechRecognition**：Python库，可以与多种语音识别服务集成

### 3. 云服务API

**方案描述**：
- 利用第三方云服务提供的说话人分割API
- 将音频发送到API服务，获取带有说话人标记的转录结果

**可选服务**：
- **Google Cloud Speech-to-Text**
- **Microsoft Azure Speech Service**
- **AWS Transcribe**

## 推荐实现路径

基于我们的需求和当前系统架构，建议采用以下实现路径：

### 短期解决方案：Whisper + PyAnnote/Resemblyzer

1. 保持当前的Whisper模型用于文字转录
2. 引入PyAnnote或Resemblyzer进行说话人分割
3. 开发模块协调两个处理流程并合并结果
4. 更新数据模型，支持说话人信息的存储和展示

### 长期解决方案：考虑迁移到集成式解决方案

1. 评估Whisper-Diarization等集成式解决方案的性能和准确性
2. 如果性能良好，考虑将其作为默认转录引擎
3. 提供选项让用户选择转录引擎，兼顾不同场景需求

## 技术实现细节

### 文件结构

```
backend/
  services/
    whisper.go (现有)
    diarization.go (新增)
    transcription.go (整合whisper和diarization)
```

### 接口设计

```go
// 说话人分割服务
type DiarizationService interface {
    // 执行说话人分割，返回包含时间戳和发言人ID的切片
    DiarizeAudio(audioData []byte) ([]SpeakerSegment, error)
}

// 说话人分割段落
type SpeakerSegment struct {
    Start     float64 // 开始时间(秒)
    End       float64 // 结束时间(秒)
    SpeakerID string  // 说话人标识
}

// 集成说话人信息的转录服务
type TranscriptionService interface {
    // 生成包含说话人信息的完整转录
    TranscribeWithSpeakers(audioData []byte) ([]TranscriptSegment, error)
}

// 转录段落
type TranscriptSegment struct {
    Start     float64 // 开始时间(秒)
    End       float64 // 结束时间(秒)
    SpeakerID string  // 说话人标识
    Text      string  // 转录文本
}
```

### 处理流程

1. 接收音频文件
2. 并行处理：
   - Whisper进行文字转录
   - 说话人分割服务进行说话人识别
3. 合并两个结果，根据时间戳对齐
4. 返回包含说话人信息的完整转录结果
5. 前端展示时区分不同说话人的内容

## 用户体验设计

### 前端展示

- 不同说话人的内容使用不同颜色或标签区分
- 提供说话人筛选功能，可以只查看特定人员的发言
- 允许用户编辑和修正自动识别的说话人信息

### 数据导出

- 导出的会议纪要中保留说话人信息
- 支持将说话人信息同步到Notion等第三方服务

## 预期挑战

1. **准确性**：说话人分割技术在嘈杂环境或多人快速交谈场景下准确率可能不高
2. **资源消耗**：添加说话人分割会增加计算资源消耗
3. **延迟**：处理时间可能增加
4. **跨语言支持**：不同语言的支持程度可能不同

## 后续步骤

1. 评估并选择说话人分割库/服务
2. 开发原型验证技术可行性
3. 集成到现有系统并进行测试
4. 优化性能和准确性
5. 更新文档和用户指南 