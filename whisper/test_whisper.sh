#!/bin/bash

echo "测试Whisper集成"

# 获取当前脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# 确保使用Anaconda环境中的Python
PYTHON_CMD="/Users/xujiawei/anaconda3/bin/python"

if [ ! -f "$PYTHON_CMD" ]; then
  echo "错误: 找不到指定的Python解释器: $PYTHON_CMD"
  exit 1
fi

echo "检查whisper模块是否已安装..."
$PYTHON_CMD -c "import whisper; print('Whisper模块已安装')" || {
  echo "错误: Whisper模块未安装，请运行: pip install openai-whisper"
  exit 1
}

echo "测试转录功能..."
$PYTHON_CMD "$SCRIPT_DIR/whisper_test.py" --help

echo "测试完成"
exit 0 