# 🧪 AppWright 测试文件

这个目录包含用于 BrowserStack App Automate 测试的 AppWright 测试文件。

## 📁 文件说明

### `login.spec.js`
- **功能**: 登录流程测试
- **测试场景**: 
  - 有效凭据登录
  - 无效凭据错误处理
  - 表单验证

### `onboarding.spec.js`
- **功能**: 新用户引导流程测试
- **测试场景**:
  - 完整引导流程
  - 跳过可选步骤
  - 表单验证

### `e2e.spec.js`
- **功能**: 端到端用户旅程测试
- **测试场景**:
  - 从注册到购买的完整流程
  - 支付失败恢复

## 🚀 如何使用

### 1. 提交测试任务
```bash
./qgjob submit \
  --org-id=your-org \
  --app-version-id=bs://your-app-id \
  --test=tests/login.spec.js \
  --target=browserstack
```

### 2. 查看测试结果
```bash
./qgjob status --job-id=<job-id>
```

### 3. 批量运行测试
```bash
# 运行所有测试
for test in tests/*.spec.js; do
  ./qgjob submit \
    --org-id=your-org \
    --app-version-id=bs://your-app-id \
    --test="$test" \
    --target=browserstack
done
```

## 📝 编写新测试

### 基本结构
```javascript
import { test, expect } from '@playwright/test';

test.describe('Your Test Suite', () => {
  test('Your test case', async ({ page }) => {
    // 测试步骤
    await page.goto('https://your-app.com');
    
    // 断言
    await expect(page.locator('[data-testid="element"]')).toBeVisible();
  });
});
```

### 最佳实践
1. **使用 data-testid**: 选择器更稳定
2. **描述性测试名称**: 清楚说明测试目的
3. **独立测试**: 每个测试不依赖其他测试
4. **错误处理**: 包含错误场景测试
5. **清理**: 测试后清理状态

## 🔧 配置

### BrowserStack 设备配置
在 `internal/agent/appwright_agent.go` 中配置：

```go
"devices": []map[string]interface{}{
    {
        "deviceName": "iPhone 14",
        "osVersion":  "16",
        "projectName": "QualGent Test Platform",
        "buildName":   fmt.Sprintf("build-%s", job.AppVersionId),
        "sessionName": fmt.Sprintf("test-%s", job.TestPath),
    },
}
```

### 环境变量
```bash
export BROWSERSTACK_USERNAME=your_username
export BROWSERSTACK_ACCESS_KEY=your_access_key
```

## 📊 测试报告

测试完成后，可以通过以下方式查看结果：

1. **CLI 输出**: `./qgjob status --job-id=<job-id>`
2. **JSON 格式**: `./qgjob status --job-id=<job-id> --json`
3. **BrowserStack 仪表板**: 使用 session_id 查看详细结果
4. **视频录制**: 查看测试执行视频
5. **日志分析**: 查看详细错误信息

## 🎯 下一步

1. **自定义测试**: 根据你的应用修改测试用例
2. **添加更多测试**: 覆盖更多功能场景
3. **CI/CD 集成**: 在 GitHub Actions 中自动运行测试
4. **性能测试**: 添加性能相关的测试用例 