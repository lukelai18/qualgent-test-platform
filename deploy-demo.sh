#!/bin/bash

echo "🚀 部署演示应用到 GitHub Pages"

# 检查是否在 git 仓库中
if [ ! -d ".git" ]; then
    echo "❌ 当前目录不是 git 仓库"
    exit 1
fi

# 创建 gh-pages 分支
echo "📦 创建 gh-pages 分支..."
git checkout -b gh-pages 2>/dev/null || git checkout gh-pages

# 复制演示应用文件
echo "📁 复制演示应用文件..."
cp -r demo-app/* .

# 添加文件到 git
git add .
git commit -m "Deploy demo app for testing" 2>/dev/null || echo "No changes to commit"

# 推送到远程仓库
echo "🚀 推送到远程仓库..."
git push origin gh-pages

echo "✅ 部署完成！"
echo "🌐 你的演示应用将在以下 URL 可用："
echo "   https://[your-username].github.io/[your-repo-name]/"
echo ""
echo "📝 请将上面的 URL 替换到测试文件中，然后运行测试！" 