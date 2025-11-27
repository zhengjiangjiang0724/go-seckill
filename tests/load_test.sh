#!/bin/bash

# 压力测试脚本
# 使用Apache Bench (ab) 或 wrk 进行压力测试

BASE_URL="http://localhost:8080/api/v1"
PRODUCT_ID=1

echo "=== Go秒杀系统压力测试 ==="
echo ""

# 测试1: 获取商品列表
echo "1. 测试获取商品列表接口..."
ab -n 10000 -c 100 "${BASE_URL}/products" > load_test_products.log 2>&1
echo "完成，结果保存在 load_test_products.log"

# 测试2: 生成令牌
echo ""
echo "2. 测试生成令牌接口..."
for i in {1..1000}; do
    curl -X POST "${BASE_URL}/seckill/token" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"user_${i}\",\"product_id\":${PRODUCT_ID}}" \
        -w "%{http_code}\n" -o /dev/null -s &
    
    # 控制并发数
    if (( i % 100 == 0 )); then
        wait
    fi
done
wait
echo "完成"

# 测试3: 秒杀接口（需要先获取令牌）
echo ""
echo "3. 测试秒杀接口..."
TOKEN_FILE="/tmp/seckill_tokens.txt"
rm -f "$TOKEN_FILE"

# 先生成一批令牌
for i in {1..500}; do
    TOKEN=$(curl -X POST "${BASE_URL}/seckill/token" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"user_${i}\",\"product_id\":${PRODUCT_ID}}" \
        -s | jq -r '.data.token')
    
    if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
        echo "${TOKEN}" >> "$TOKEN_FILE"
    fi
done

# 使用令牌进行秒杀
TOKEN_COUNT=$(wc -l < "$TOKEN_FILE")
echo "生成了 $TOKEN_COUNT 个令牌，开始秒杀..."

while read -r TOKEN; do
    curl -X POST "${BASE_URL}/seckill/buy" \
        -H "Content-Type: application/json" \
        -d "{\"user_id\":\"user_$(date +%s)\",\"product_id\":${PRODUCT_ID},\"token\":\"${TOKEN}\"}" \
        -w "%{http_code}\n" -o /dev/null -s &
    
    if (( $(jobs -r | wc -l) >= 50 )); then
        wait
    fi
done < "$TOKEN_FILE"
wait

echo "完成"

echo ""
echo "=== 压力测试完成 ==="
echo "详细日志请查看 load_test_products.log"

