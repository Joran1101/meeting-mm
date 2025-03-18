#!/usr/bin/env python
# -*- coding: utf-8 -*-

import sys
import os
import whisper
import argparse
import torch

def main():
    parser = argparse.ArgumentParser(description="使用OpenAI Whisper模型转录音频")
    parser.add_argument("audio_file", help="要转录的音频文件路径")
    parser.add_argument("--model", default="base", help="要使用的模型 (tiny, base, small, medium, large)")
    args = parser.parse_args()

    if not os.path.exists(args.audio_file):
        print(f"错误：文件 {args.audio_file} 不存在", file=sys.stderr)
        return 1

    # 强制使用CPU，并设置为Float32精度
    device = "cpu"
    torch.set_default_tensor_type(torch.FloatTensor)
    
    print(f"加载模型 {args.model}...", file=sys.stderr)
    model = whisper.load_model(args.model, device=device)

    print("加载音频...", file=sys.stderr)
    audio = whisper.load_audio(args.audio_file)
    audio = whisper.pad_or_trim(audio)

    print("生成梅尔频谱图...", file=sys.stderr)
    mel = whisper.log_mel_spectrogram(audio, n_mels=model.dims.n_mels).to(device)

    print("检测语言...", file=sys.stderr)
    _, probs = model.detect_language(mel)
    detected_language = max(probs, key=probs.get)
    print(f"检测到的语言: {detected_language} (置信度: {probs[detected_language]:.2%})", file=sys.stderr)

    print("转录音频...", file=sys.stderr)
    options = whisper.DecodingOptions(fp16=False)  # 禁用fp16
    result = whisper.decode(model, mel, options)

    print("\n转录结果:")
    print(result.text)
    
    return 0

if __name__ == "__main__":
    sys.exit(main()) 