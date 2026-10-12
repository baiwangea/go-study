#!/bin/bash

echo "════════════════════════════════════════════════"
echo "  测试所有重构后的学习文件"
echo "════════════════════════════════════════════════"

total=0
success=0
failed=0

# 01-go-fundamentals
echo -e "\n【01-go-fundamentals】"
for f in 01-go-fundamentals/functions/*.go 01-go-fundamentals/packages/*.go 01-go-fundamentals/interfaces/*.go; do
  [ -f "$f" ] || continue
  head -1 "$f" | grep -q "//go:build level" || continue
  total=$((total + 1))
  echo -n "  测试 $(basename $f)... "
  if go run "$f" > /dev/null 2>&1; then
    echo "✅"
    success=$((success + 1))
  else
    echo "❌"
    failed=$((failed + 1))
  fi
done

# 02-go-pointers
echo -e "\n【02-go-pointers】"
for f in 02-go-pointers/1*.go; do
  [ -f "$f" ] || continue
  total=$((total + 1))
  echo -n "  测试 $(basename $f)... "
  if go run "$f" > /dev/null 2>&1; then
    echo "✅"
    success=$((success + 1))
  else
    echo "❌"
    failed=$((failed + 1))
  fi
done

# 03-go-concurrency
echo -e "\n【03-go-concurrency】"
for dir in goroutines channels selects contextpkg patterns mutexes; do
  for f in 03-go-concurrency/$dir/*.go; do
    [ -f "$f" ] || continue
    head -1 "$f" | grep -q "//go:build level" || continue
    total=$((total + 1))
    echo -n "  测试 $dir/$(basename $f)... "
    if timeout 3 go run "$f" > /dev/null 2>&1; then
      echo "✅"
      success=$((success + 1))
    else
      echo "❌"
      failed=$((failed + 1))
    fi
  done
done

# 04-goframe (跳过需要外部依赖的)
echo -e "\n【04-goframe】(跳过需要数据库的关卡)"
for f in 04-goframe/0{1,2,3,4,5,6}*.go 04-goframe/1{0,1,2,4,5,6}*.go; do
  [ -f "$f" ] || continue
  head -1 "$f" | grep -q "//go:build level" || continue
  total=$((total + 1))
  echo -n "  测试 $(basename $f)... "
  if timeout 5 go run "$f" > /dev/null 2>&1; then
    echo "✅"
    success=$((success + 1))
  else
    echo "⚠️ (可能需要外部依赖)"
    failed=$((failed + 1))
  fi
done

echo -e "\n════════════════════════════════════════════════"
echo "  总计: $total 个文件"
echo "  成功: $success ✅"
echo "  失败: $failed ❌"
echo "  成功率: $(( success * 100 / total ))%"
echo "════════════════════════════════════════════════"
