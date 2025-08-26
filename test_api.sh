#!/bin/bash

# API测试脚本
BASE_URL="http://localhost:8080"

echo "=== 测试扩展API接口 ==="

# 1. 注册用户
echo "1. 注册用户..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }')
echo "注册响应: $REGISTER_RESPONSE"

# 2. 登录获取token
echo -e "\n2. 用户登录..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }')
echo "登录响应: $LOGIN_RESPONSE"

# 提取token
TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Token: $TOKEN"

if [ -z "$TOKEN" ]; then
  echo "登录失败，无法获取token"
  exit 1
fi

# 3. 创建分类
echo -e "\n3. 创建分类..."
CATEGORY_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/categories/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "工作",
    "color": "#FF5722",
    "icon": "work"
  }')
echo "创建分类响应: $CATEGORY_RESPONSE"

# 4. 获取分类列表
echo -e "\n4. 获取分类列表..."
GET_CATEGORIES_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/categories" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN")
echo "分类列表响应: $GET_CATEGORIES_RESPONSE"

# 5. 创建扩展TODO
echo -e "\n5. 创建扩展TODO..."
TODO_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/todos/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "学习Go语言",
    "description": "学习Go语言基础语法和框架",
    "priority": 2,
    "due_date": "2023-12-31T23:59:59Z",
    "tags": ["学习", "编程", "Go"],
    "category_id": 1,
    "reminder": "2023-12-30T09:00:00Z"
  }')
echo "创建TODO响应: $TODO_RESPONSE"

# 6. 获取扩展TODO列表
echo -e "\n6. 获取扩展TODO列表..."
GET_TODOS_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/todos/list" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "limit": 10,
    "offset": 0
  }')
echo "TODO列表响应: $GET_TODOS_RESPONSE"

# 7. 搜索TODO
echo -e "\n7. 搜索TODO..."
SEARCH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/todos/search" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "keyword": "Go",
    "limit": 10,
    "offset": 0
  }')
echo "搜索响应: $SEARCH_RESPONSE"

# 8. 获取用户设置
echo -e "\n8. 获取用户设置..."
GET_SETTINGS_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/settings" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN")
echo "用户设置响应: $GET_SETTINGS_RESPONSE"

# 9. 更新用户设置
echo -e "\n9. 更新用户设置..."
UPDATE_SETTINGS_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v2/settings/update" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "theme": "dark",
    "notification_time": "08:00",
    "language": "en-US",
    "timezone": "America/New_York"
  }')
echo "更新设置响应: $UPDATE_SETTINGS_RESPONSE"

echo -e "\n=== API测试完成 ==="