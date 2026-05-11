#!/bin/bash
# Claude Code PostToolUse hook: 根据编辑的文件路径自动运行格式化工具
# 前端 (frontend/): pnpm lint --fix   后端 (backend/): gofumpt + goimports

set -euo pipefail

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
if [ -z "$file_path" ]; then
  exit 0
fi

PROJECT_DIR="/Users/yuyong/personal_proj/wetalk"

# file_path 是绝对路径 (Edit/Write 工具传入)
if [[ "$file_path" == "$PROJECT_DIR/frontend/"* ]] && [[ "$file_path" =~ \.(ts|tsx|js|jsx|css)$ ]]; then
  cd "$PROJECT_DIR/frontend"
  pnpm lint --fix "$file_path" 2>&1 || true
elif [[ "$file_path" == "$PROJECT_DIR/backend/"* ]] && [[ "$file_path" =~ \.go$ ]]; then
  gofumpt -w "$file_path" 2>&1 || true
  goimports -w "$file_path" 2>&1 || true
fi
