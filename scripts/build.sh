# 变量 版本号
VERSION=$1
if [ -z "$VERSION" ]; then
    VERSION=0.1.4
fi
BUILD_TIME=$(date +%Y-%m-%d\ %H:%M:%S)

# dist/${VERSION}/ 清空文件夹
rm -rf dist/${VERSION}/

# 复制配置文件
mkdir -p dist/${VERSION}/configs
cp -r configs/ dist/${VERSION}/

# 同步代码中的构建版本和构建时间
sed -i "s/const BuildVersion = \".*\"/const BuildVersion = \"${VERSION}\"/g" internal/consts/consts.go
sed -i "s/const BuildVersion = \".*\"/const BuildVersion = \"${VERSION}\"/g" resource/client/index.html
sed -i "s/const BuildTime = \".*\"/const BuildTime = \"${BUILD_TIME}\"/g" internal/consts/consts.go
sed -i "s/const BuildTime = \".*\"/const BuildTime = \"${BUILD_TIME}\"/g" resource/client/index.html

# # 编译
# 可选: 平台 linux windows darwin   架构 amd64 arm64
# 安装Xgo: go install src.techknowlogick.com/xgo@latest
xgo -go go-1.22.10 -out dist/${VERSION}/wp -targets=linux/amd64,linux/arm64,windows/amd64 -ldflags="-w -s -X 'main.Version=${VERSION}'" .