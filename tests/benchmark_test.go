package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"go-seckill/models"
)

const (
	baseURL = "http://localhost:8080/api/v1"
)

// 测试数据
var (
	testProductID uint
	testUserIDs   = []string{
		"user1", "user2", "user3", "user4", "user5",
		"user6", "user7", "user8", "user9", "user10",
	}
	tokens = make(map[string]string)
)

// TestCreateProduct 创建测试商品
func TestCreateProduct(t *testing.T) {
	product := models.Product{
		Name:          "测试秒杀商品",
		Price:         99.99,
		Stock:         10000,
		SeckillStock:  1000,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
	}

	body, _ := json.Marshal(product)
	resp, err := http.Post(baseURL+"/admin/products", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if data, ok := result["data"].(map[string]interface{}); ok {
		if id, ok := data["id"].(float64); ok {
			testProductID = uint(id)
			t.Logf("Created product with ID: %d", testProductID)
		}
	}
}

// TestGenerateTokens 生成令牌
func TestGenerateTokens(t *testing.T) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, userID := range testUserIDs {
		wg.Add(1)
		go func(uid string) {
			defer wg.Done()

			reqBody := map[string]interface{}{
				"user_id":   uid,
				"product_id": testProductID,
			}
			body, _ := json.Marshal(reqBody)

			resp, err := http.Post(baseURL+"/seckill/token", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Logf("Failed to generate token for user %s: %v", uid, err)
				return
			}
			defer resp.Body.Close()

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return
			}

			if data, ok := result["data"].(map[string]interface{}); ok {
				if token, ok := data["token"].(string); ok {
					mu.Lock()
					tokens[uid] = token
					mu.Unlock()
				}
			}
		}(userID)
	}

	wg.Wait()
	t.Logf("Generated %d tokens", len(tokens))
}

// BenchmarkSeckill 性能测试：并发秒杀
func BenchmarkSeckill(b *testing.B) {
	if testProductID == 0 {
		b.Skip("Product not created")
	}

	// 预热：生成令牌
	TestGenerateTokens(&testing.T{})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		userIndex := 0
		for pb.Next() {
			userID := fmt.Sprintf("user_%d", userIndex%1000)
			userIndex++

			// 生成令牌
			reqBody := map[string]interface{}{
				"user_id":    userID,
				"product_id": testProductID,
			}
			body, _ := json.Marshal(reqBody)

			resp, err := http.Post(baseURL+"/seckill/token", "application/json", bytes.NewBuffer(body))
			if err != nil {
				continue
			}

			var tokenResult map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&tokenResult)
			resp.Body.Close()

			var token string
			if data, ok := tokenResult["data"].(map[string]interface{}); ok {
				if t, ok := data["token"].(string); ok {
					token = t
				}
			}

			if token == "" {
				continue
			}

			// 执行秒杀
			seckillReq := map[string]interface{}{
				"user_id":    userID,
				"product_id": testProductID,
				"token":      token,
			}
			seckillBody, _ := json.Marshal(seckillReq)

			resp, err = http.Post(baseURL+"/seckill/buy", "application/json", bytes.NewBuffer(seckillBody))
			if err != nil {
				continue
			}
			resp.Body.Close()
		}
	})
}

// TestConcurrentSeckill 并发秒杀测试
func TestConcurrentSeckill(t *testing.T) {
	if testProductID == 0 {
		t.Skip("Product not created")
	}

	concurrency := 100
	requestsPerGoroutine := 10
	var successCount int64
	var failCount int64
	var mu sync.Mutex

	startTime := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			userID := fmt.Sprintf("concurrent_user_%d", id)
			for j := 0; j < requestsPerGoroutine; j++ {
				// 生成令牌
				reqBody := map[string]interface{}{
					"user_id":    userID,
					"product_id": testProductID,
				}
				body, _ := json.Marshal(reqBody)

				resp, err := http.Post(baseURL+"/seckill/token", "application/json", bytes.NewBuffer(body))
				if err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					continue
				}

				var tokenResult map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&tokenResult)
				resp.Body.Close()

				var token string
				if data, ok := tokenResult["data"].(map[string]interface{}); ok {
					if t, ok := data["token"].(string); ok {
						token = t
					}
				}

				if token == "" {
					mu.Lock()
					failCount++
					mu.Unlock()
					continue
				}

				// 执行秒杀
				seckillReq := map[string]interface{}{
					"user_id":    userID,
					"product_id": testProductID,
					"token":      token,
				}
				seckillBody, _ := json.Marshal(seckillReq)

				resp, err = http.Post(baseURL+"/seckill/buy", "application/json", bytes.NewBuffer(seckillBody))
				if err != nil {
					mu.Lock()
					failCount++
					mu.Unlock()
					continue
				}

				if resp.StatusCode == 200 {
					mu.Lock()
					successCount++
					mu.Unlock()
				} else {
					mu.Lock()
					failCount++
					mu.Unlock()
				}
				resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	t.Logf("Concurrent Seckill Test Results:")
	t.Logf("  Total Requests: %d", concurrency*requestsPerGoroutine)
	t.Logf("  Success: %d", successCount)
	t.Logf("  Failed: %d", failCount)
	t.Logf("  Duration: %v", duration)
	t.Logf("  QPS: %.2f", float64(successCount)/duration.Seconds())
	t.Logf("  Average Latency: %v", duration/time.Duration(concurrency*requestsPerGoroutine))
}

