-- wrk 性能测试脚本
-- 使用方法: wrk -t4 -c100 -d30s -s wrk_test.lua http://localhost:8080

-- 初始化
init = function(args)
    -- 商品ID（需要根据实际情况修改）
    product_id = 1
    user_counter = 0
end

-- 请求生成
request = function()
    user_counter = user_counter + 1
    user_id = "wrk_user_" .. user_counter
    
    -- 生成令牌请求
    wrk.method = "POST"
    wrk.body = string.format('{"user_id":"%s","product_id":%d}', user_id, product_id)
    wrk.headers["Content-Type"] = "application/json"
    
    return wrk.format("POST", "/api/v1/seckill/token")
end

-- 响应处理
response = function(status, headers, body)
    -- 可以在这里解析响应并提取token，用于后续的秒杀请求
    -- 为了简化，这里只返回状态
    if status == 200 then
        -- 尝试提取token
        local token = string.match(body, '"token":"([^"]+)"')
        if token then
            -- 可以保存token用于后续请求
        end
    end
end

