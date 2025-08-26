#!/bin/bash

# 测试数据同步API接口
# 使用方法: ./test_sync_api.sh

BASE_URL="http://127.0.0.1:8080"
TOKEN=""

echo "=== 测试数据同步API接口 ==="

# 1. 用户注册
echo "1. 注册测试用户..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "synctest",
    "email": "synctest@example.com", 
    "password": "password123"
  }')
echo "注册响应: $REGISTER_RESPONSE"

# 2. 用户登录获取token
echo -e "\n2. 用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "synctest",
    "password": "password123"
  }')
echo "登录响应: $LOGIN_RESPONSE"

# 提取token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    echo "错误: 无法获取token"
    exit 1
fi
echo "获取到token: $TOKEN"

# 3. 创建一些测试数据
echo -e "\n3. 创建测试TODO..."
CREATE_TODO_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/todos/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "同步测试任务",
    "description": "用于测试数据同步功能",
    "priority": 1,
    "tags": ["测试", "同步"]
  }')
echo "创建TODO响应: $CREATE_TODO_RESPONSE"

echo -e "\n4. 创建测试分类..."
CREATE_CATEGORY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/categories/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "同步测试分类",
    "color": "#FF5722",
    "icon": "test"
  }')
echo "创建分类响应: $CREATE_CATEGORY_RESPONSE"

# 5. 测试获取同步版本
echo -e "\n5. 获取同步版本..."
SYNC_VERSION_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/sync/version" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{}')
echo "同步版本响应: $SYNC_VERSION_RESPONSE"

# 6. 测试增量同步
echo -e "\n6. 测试增量同步..."
INCREMENTAL_SYNC_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/sync/todos" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "since": 0
  }')
echo "增量同步响应: $INCREMENTAL_SYNC_RESPONSE"

# 7. 测试批量同步
echo -e "\n7. 测试批量同步..."
BATCH_SYNC_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/sync/batch" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "todos": [
      {
        "id": 0,
        "title": "批量同步测试任务",
        "description": "通过批量同步创建的任务",
        "completed": false,
        "priority": 2,
        "tags": ["批量", "同步"],
        "is_deleted": false,
        "sync_version": 0,
        "updated_at": "2023-01-01T00:00:00Z"
      }
    ],
    "categories": [
      {
        "id": 0,
        "name": "批量同步分类",
        "color": "#4CAF50",
        "icon": "batch",
        "is_deleted": false,
        "sync_version": 0,
        "updated_at": "2023-01-01T00:00:00Z"
      }
    ]
  }')
echo "批量同步响应: $BATCH_SYNC_RESPONSE"

echo -e "\n=== 测试完成 ==="