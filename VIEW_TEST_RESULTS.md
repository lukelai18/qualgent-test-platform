# 📊 如何查看测试结果

## 🎯 **测试结果查看方法**

### 1. **基本状态查看**
```bash
./qgjob status --job-id=<job-id>
```

**输出示例：**
```
Job Status:
Job ID: 07a5e68d-bcc0-453b-9593-62c5014b83a9
Status: COMPLETED
Created: 2025-07-20T06:39:47Z
Completed: 2025-07-20T06:39:51Z
Session ID: demo-session-123
Logs URL: https://app-automate.browserstack.com/dashboard/v2/builds/demo-session-123/sessions/demo-session-123
Video URL: https://app-automate.browserstack.com/dashboard/v2/builds/demo-session-123/sessions/demo-session-123/video
Duration: 45 seconds
```

### 2. **JSON 格式输出**
```bash
./qgjob status --job-id=<job-id> --json
```

**输出示例：**
```json
{
  "job_id": "07a5e68d-bcc0-453b-9593-62c5014b83a9",
  "status": "COMPLETED",
  "created_at": "2025-07-20T06:39:47Z",
  "completed_at": "2025-07-20T06:39:51Z",
  "session_id": "demo-session-123",
  "logs_url": "https://app-automate.browserstack.com/dashboard/v2/builds/demo-session-123/sessions/demo-session-123",
  "video_url": "https://app-automate.browserstack.com/dashboard/v2/builds/demo-session-123/sessions/demo-session-123/video",
  "test_duration": 45,
  "error_message": ""
}
```

### 3. **数据库直接查询**
```bash
docker-compose exec postgres psql -U user -d qg_jobs -c "
SELECT 
  id, 
  status, 
  session_id, 
  logs_url, 
  video_url, 
  test_duration, 
  error_message,
  created_at,
  completed_at
FROM jobs 
WHERE id = '<job-id>';
"
```

### 4. **查看测试日志**
- **BrowserStack 日志**: 打开 `logs_url` 链接查看详细测试日志
- **视频录制**: 打开 `video_url` 链接查看测试执行视频
- **会话详情**: 使用 `session_id` 在 BrowserStack 仪表板中查找

## 📋 **测试结果字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| `job_id` | string | 任务唯一标识符 |
| `status` | string | 任务状态 (PENDING/SCHEDULED/ASSIGNED/RUNNING/COMPLETED/FAILED) |
| `created_at` | timestamp | 任务创建时间 |
| `completed_at` | timestamp | 任务完成时间 |
| `session_id` | string | BrowserStack 会话 ID |
| `logs_url` | string | 测试日志查看链接 |
| `video_url` | string | 测试视频查看链接 |
| `test_duration` | int | 测试执行时长（秒） |
| `error_message` | string | 错误信息（如果有） |

## 🔍 **状态生命周期**

1. **PENDING**: 任务已提交，等待调度
2. **SCHEDULED**: 任务已分组，等待分配
3. **ASSIGNED**: 任务已分配给 Agent
4. **RUNNING**: 测试正在执行
5. **COMPLETED**: 测试成功完成
6. **FAILED**: 测试执行失败

## 🛠️ **实用命令**

### 查看所有任务
```bash
docker-compose exec postgres psql -U user -d qg_jobs -c "
SELECT id, status, created_at, completed_at 
FROM jobs 
ORDER BY created_at DESC 
LIMIT 10;
"
```

### 查看失败的任务
```bash
docker-compose exec postgres psql -U user -d qg_jobs -c "
SELECT id, status, error_message, created_at 
FROM jobs 
WHERE status = 'FAILED' 
ORDER BY created_at DESC;
"
```

### 查看测试统计
```bash
docker-compose exec postgres psql -U user -d qg_jobs -c "
SELECT 
  status,
  COUNT(*) as count,
  AVG(test_duration) as avg_duration
FROM jobs 
GROUP BY status;
"
```

## 🎯 **BrowserStack 集成**

### 真实测试结果查看
1. **设置环境变量**:
   ```bash
   export BROWSERSTACK_USERNAME=your_username
   export BROWSERSTACK_ACCESS_KEY=your_access_key
   ```

2. **提交真实测试**:
   ```bash
   ./qgjob submit \
     --org-id=your-org \
     --app-version-id=bs://your-app-id \
     --test=your-test.spec.js \
     --target=browserstack
   ```

3. **查看结果**:
   ```bash
   ./qgjob status --job-id=<job-id>
   ```

### 访问 BrowserStack 仪表板
- **App Automate**: https://app-automate.browserstack.com/dashboard
- **会话详情**: 使用 `session_id` 在仪表板中查找
- **实时监控**: 查看测试执行过程
- **日志分析**: 查看详细错误信息

## 📊 **CI/CD 集成**

### GitHub Actions 示例
```yaml
name: AppWright Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Submit test
        run: |
          ./qgjob submit \
            --org-id=${{ secrets.ORG_ID }} \
            --app-version-id=${{ secrets.APP_VERSION_ID }} \
            --test=tests/e2e.spec.js \
            --target=browserstack
      - name: Wait and check results
        run: |
          sleep 30
          ./qgjob status --job-id=<job-id> --json
```

## 🚀 **高级功能**

### 批量结果查询
```bash
# 获取所有完成的任务
docker-compose exec postgres psql -U user -d qg_jobs -c "
SELECT id, status, session_id, test_duration 
FROM jobs 
WHERE status = 'COMPLETED' 
ORDER BY completed_at DESC;
"
```

### 性能分析
```bash
# 查看测试执行时间统计
docker-compose exec postgres psql -U user -d qg_jobs -c "
SELECT 
  AVG(test_duration) as avg_duration,
  MIN(test_duration) as min_duration,
  MAX(test_duration) as max_duration,
  COUNT(*) as total_tests
FROM jobs 
WHERE status = 'COMPLETED' AND test_duration > 0;
"
```

---

**🎉 现在你可以完整地查看和管理测试结果了！** 