# 拉取镜像

docker pull ubuntu

# 重命名


docker tag ubuntu test.com/ubuntu

# 推送

docker push test.com/ubuntu

# 保存镜像

docker save -o ubuntu.tar ubuntu

# 解压镜像


docker load -i ubuntu.tar