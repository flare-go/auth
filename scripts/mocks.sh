#!/bin/bash

# 定義源碼目錄（根目錄）
SRC_DIR="."

# 定義輸出目錄
MOCK_DIR="./mocks"

# 創建輸出目錄（如果不存在）
mkdir -p $MOCK_DIR

# 遍歷源碼目錄
find $SRC_DIR -name "*.go" | while read file; do
    # 跳過 mocks 目錄和 vendor 目錄
    if [[ $file == *"/mocks/"* ]] || [[ $file == *"/vendor/"* ]]; then
        continue
    fi

    # 獲取相對路徑
    rel_path=${file#$SRC_DIR/}
    dir_path=$(dirname $rel_path)

    # 提取文件名（不帶擴展名）
    filename=$(basename "$file" .go)

    # 提取包名
    package=$(grep "^package" "$file" | awk '{print $2}')

    # 生成 mock 文件名
    mock_file="$MOCK_DIR/mock_${package}_${filename}.go"

    # 檢查文件是否包含 interface 定義
    if grep -q "type.*interface" "$file"; then
        echo "Generating mock for $file"
        mockgen -source="$file" -destination="$mock_file" -package=mocks
    fi
done

echo "Mock generation complete."