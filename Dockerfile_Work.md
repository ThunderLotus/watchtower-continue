# 📋 Dockerfile 版本管理使用指南

##  🎯 当前实现

  您的 Dockerfile 现在使用 ARG WATCHTOWER_VERSION 参数来动态设置版本号：

   \# 定义版本参数（默认值为 main）
   ARG WATCHTOWER_VERSION=main

   \# 构建时使用该参数
   RUN ... -ldflags "-X github.com/containrrr/watchtower/internal/meta.Version=${WATCHTOWER_VERSION}"

  ---

 ## 💡 使用方法

  ### 方法 1：使用默认版本（从 Go 代码读取）

   \# 不指定版本，使用 Dockerfile 中的默认值（main）
   docker build . -f dockerfiles/Dockerfile.dev-self-contained -t watchtower

  注意：此时版本号会从 internal/meta/meta.go 中读取：
   Version = "v1.8.0"  // 实际使用的版本

 ###  方法 2：指定版本号

   \# 指定具体版本
   docker build . -f dockerfiles/Dockerfile.dev-self-contained \
     --build-arg WATCHTOWER_VERSION=v1.8.0 \
     -t watchtower:v1.8.0

   \# 指定分支
   docker build . -f dockerfiles/Dockerfile.dev-self-contained \
     --build-arg WATCHTOWER_VERSION=main \
     -t watchtower:latest

  ### 方法 3：使用环境变量（推荐用于 CI/CD）

   \# 设置环境变量
   export WATCHTOWER_VERSION=v1.8.0

   \# 使用环境变量构建
   docker build . -f dockerfiles/Dockerfile.dev-self-contained \
     --build-arg WATCHTOWER_VERSION=${WATCHTOWER_VERSION} \
     -t watchtower:${WATCHTOWER_VERSION}

  ### 方法 4：多架构构建

   \# 同时构建多个架构
   docker buildx build . -f dockerfiles/Dockerfile.dev-self-contained \
     --build-arg WATCHTOWER_VERSION=v1.8.0 \
     --platform linux/amd64,linux/arm64 \
     -t watchtower:v1.8.0 \
     --push

  ---

##  ⚠️ 注意事项

###  1. 版本号格式

  支持的格式：
   - ✅ v1.8.0 - 语义化版本
   - ✅ main - 主分支
   - ✅ develop - 开发分支
   - ✅ docker-upgrade - 功能分支
   - ✅ abc123 - Git 提交 SHA

  不支持的格式：
   - ❌ 1.8.0 - 缺少 v 前缀（可能导致混淆）
   - ❌ version-1.8.0 - 非标准格式

###  2. 与 Go 代码的一致性

  重要：Docker 构建的版本号必须与 Go 代码中的版本号保持一致。

  不一致的后果：
   \# Dockerfile 中指定 v1.9.0
   docker build ... --build-arg WATCHTOWER_VERSION=v1.9.0

   \# 但 Go 代码中还是 v1.8.0
   \# internal/meta/meta.go
   Version = "v1.8.0"

   \# 结果：显示的版本和实际功能可能不匹配

  最佳实践：
   \# 从 Go 代码中提取版本号
   VERSION=$(grep 'Version = ' internal/meta/meta.go | awk -F'"' '{print $2}')

   \# 使用提取的版本号构建
   docker build . -f dockerfiles/Dockerfile.dev-self-contained \
     --build-arg WATCHTOWER_VERSION=${VERSION} \
     -t watchtower:${VERSION}

###  3. Dockerfile.self-contained 的使用

   \# 这个 Dockerfile 会从 GitHub 克隆代码
   ARG WATCHTOWER_VERSION=main

   RUN git clone --branch "${WATCHTOWER_VERSION}" \
     https://github.com/ThunderLotus/watchtower-continue.git

  注意：
   - 仅用于从远程仓库构建
   - 需要网络连接
   - 版本号必须是有效的 Git 标签或分支

###  4. CI/CD 集成

  GitHub Actions 示例：
   name: Build and Push
   on:
     push:
       tags:
         - 'v*.*.*'

   jobs:
     build:
       runs-on: ubuntu-latest
       steps:
         - name: Extract version
           run: echo "VERSION=${GITHUB_REF#refs/tags/}" >> $GITHUB_ENV

         - name: Build Docker image
           run: |
             docker build . -f dockerfiles/Dockerfile.dev-self-contained \
               --build-arg WATCHTOWER_VERSION=${{ env.VERSION }} \
               -t watchtower:${{ env.VERSION }}

  GitLab CI 示例：
   build:
     stage: build
     script:
       - VERSION=$(grep 'Version = ' internal/meta/meta.go | awk -F'"' '{print $2}')
       - docker build . -f dockerfiles/Dockerfile.dev-self-contained \
         --build-arg WATCHTOWER_VERSION=${VERSION} \
         -t watchtower:${VERSION}

###  5. 本地开发工作流

  开发阶段：
   # 使用 main 分支开发
   docker build . -f dockerfiles/Dockerfile.dev-self-contained \
     --build-arg WATCHTOWER_VERSION=main \
     -t watchtower:dev

  发布阶段：
   \# 1. 更新版本号
   \# 编辑 internal/meta/meta.go
   \# Version = "v1.8.1"

   \# 2. 构建镜像
   docker build . -f dockerfiles/Dockerfile.dev-self-contained \
     --build-arg WATCHTOWER_VERSION=v1.8.1 \
     -t watchtower:v1.8.1

   \# 3. 验证版本
   docker run --rm watchtower:v1.8.1 --version

###  6. 版本号验证

  构建后验证版本号：

   \# 方法 1：运行容器查看版本
   docker run --rm watchtower:v1.8.0 --version

   \# 方法 2：检查镜像元数据
   docker inspect watchtower:v1.8.0 | grep -i version

   \# 方法 3：查看二进制文件
   docker run --rm watchtower:v1.8.0 sh -c "echo \$(\$(/watchtower --version 2>&1 | head -1))"

  ---

##  📝 最佳实践总结

|场景|推荐做法|命令示例|
|--|--|--|
| 日常开发| 使用 main 分支| --build-arg WATCHTOWER_VERSION=main |
 |功能测试  | 使用功能分支  |--build-arg WATCHTOWER_VERSION=docker-upgrade |
   | 正式发布  | 使用语义化版本   | --build-arg WATCHTOWER_VERSION=v1.8.0    |
 | CI/CD     | 从环境变量读取  | --build-arg WATCHTOWER_VERSION=${VERSION}   |
 | 多架构   | 配合 buildx 使用 | --platform linux/amd64,linux/arm64    |

  ---

 ## 🔧 常见问题

  Q1: 忘记指定版本号会怎样？
   # 会使用 Dockerfile 中的默认值：main
   docker build . -f dockerfiles/Dockerfile.dev-self-contained -t watchtower

  Q2: 如何确保版本号一致？
   # 自动从 Go 代码提取
   VERSION=$(grep 'Version = ' internal/meta/meta.go | awk -F'"' '{print $2}')
   docker build ... --build-arg WATCHTOWER_VERSION=${VERSION}

  Q3: 可以使用 Git 提交 SHA 吗？
   # 可以，但不推荐用于生产环境
   docker build ... --build-arg WATCHTOWER_VERSION=abc123def456

  Q4: 如何回滚到旧版本？
   # 使用旧的版本号重新构建
   docker build . -f dockerfiles/Dockerfile.dev-self-contained \
     --build-arg WATCHTOWER_VERSION=v1.7.1 \
     -t watchtower:v1.7.1

  ---

 ## 🎯 推荐的工作流程

   1. 开发阶段：
      docker build . --build-arg WATCHTOWER_VERSION=main -t watchtower:dev

   2. 测试阶段：
      docker build . --build-arg WATCHTOWER_VERSION=v1.8.0-rc1 -t watchtower:rc1

   3. 发布阶段：
      # 1. 更新 Go 代码版本
      # 2. 构建镜像
      VERSION=$(grep 'Version = ' internal/meta/meta.go | awk -F'"' '{print $2}')
      docker build . --build-arg WATCHTOWER_VERSION=${VERSION} -t watchtower:${VERSION}

✦ 这样您就可以灵活地管理不同场景下的版本号了！