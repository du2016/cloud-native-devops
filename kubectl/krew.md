krew是kubectl的插件管理器

# 安装

```
yum install git -y

(
  set -x; cd "$(mktemp -d)" &&
  curl -fsSLO "https://github.com/kubernetes-sigs/krew/releases/download/v0.3.2/krew.{tar.gz,yaml}" &&
  tar zxvf krew.tar.gz &&
  ./krew-"$(uname | tr '[:upper:]' '[:lower:]')_amd64" install \
    --manifest=krew.yaml --archive=krew.tar.gz
)

export PATH="${KREW_ROOT:-$HOME/.krew}/bin:$PATH"
```

重启shell

# 子命令

- help        帮助信息
- info        查看plugin的信息
- install     安装插件
- list        查看安装的插件列表
- search      查询插件
- uninstall   卸载插件
- update      更新本地插件索引
- upgrade     升级
- version     版本