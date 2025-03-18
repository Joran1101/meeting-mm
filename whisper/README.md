# Whisper.cpp 集成

本项目使用 [Whisper.cpp](https://github.com/ggerganov/whisper.cpp) 进行本地语音识别，以减少隐私风险并提高性能。

## 安装步骤

1. 克隆 Whisper.cpp 仓库：

```bash
git clone https://github.com/ggerganov/whisper.cpp.git
cd whisper.cpp
```

2. 编译 Whisper.cpp：

```bash
make
```

3. 下载模型（选择一个适合您需求的模型）：

```bash
# 基础模型（74MB）
bash ./models/download-ggml-model.sh base

# 小型模型（142MB）
bash ./models/download-ggml-model.sh small

# 中型模型（442MB）
bash ./models/download-ggml-model.sh medium

# 大型模型（1.5GB）
bash ./models/download-ggml-model.sh large
```

4. 将编译好的可执行文件和模型复制到本项目：

```bash
# 复制可执行文件
cp main /path/to/meeting-mm/whisper/

# 复制模型
cp models/ggml-base.bin /path/to/meeting-mm/whisper/models/
```

5. 更新 `.env` 文件中的模型路径：

```
WHISPER_MODEL_PATH=../whisper/models/ggml-base.bin
```

## 测试 Whisper.cpp

您可以使用以下命令测试 Whisper.cpp 是否正常工作：

```bash
cd /path/to/meeting-mm/whisper
./main -m models/ggml-base.bin -f /path/to/audio/file.wav -otxt
```

如果一切正常，您应该会看到转录结果输出到控制台，并在同一目录下生成一个文本文件。

## 模型量化

为了提高性能，您可以对模型进行量化：

```bash
# 4位量化（更小的文件大小，更快的推理，但准确性略有下降）
./models/quantize.sh models/ggml-base.bin q4_0
```

然后更新 `.env` 文件中的模型路径：

```
WHISPER_MODEL_PATH=../whisper/models/ggml-base-q4_0.bin
``` 