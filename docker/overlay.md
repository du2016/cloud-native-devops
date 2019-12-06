docker挂载选项

```
overlay on /var/lib/docker/overlay2/7392803c53647ebc72ec54c2367b00af8aedc10f51ca304d2826e04c1c40153d/merged type overlay (rw,relatime,seclabel,lowerdir=/var/lib/docker/overlay2/l/Q46AZ2IFJ2ITL6VI5RIHFSM6IU:/var/lib/docker/overlay2/l/IZJ5KCZO6DQO4YBKR6ME7LCVEM:/var/lib/docker/overlay2/l/PLUJBFFW3YPGGWL2WMBIFC3ZR5,upperdir=/var/lib/docker/overlay2/7392803c53647ebc72ec54c2367b00af8aedc10f51ca304d2826e04c1c40153d/diff,workdir=/var/lib/docker/overlay2/7392803c53647ebc72ec54c2367b00af8aedc10f51ca304d2826e04c1c40153d/work)
```

- overlay 文件系统类型
- 挂载点 /var/lib/docker/overlay2/7392803c53647ebc72ec54c2367b00af8aedc10f51ca304d2826e04c1c40153d/merged 
- rw 读写
- relatime 只有当mtime/ctime的时间戳晚于atime的时候才去更新atime
- seclabel 文件系统使用xattrs作为标签，并通过设置xattrs支持标签更改，目的是为了支持selinux
- lowerdir是您要放置新文件系统的目录，如果存在重复，则这些副本将被upperdir的版本覆盖（实际上，被隐藏）
- upperdir是您要覆盖lowerdir的目录。如果lowerdir和upperdir中存在重复的文件名，则upperdir的版本优先。
- workdir用于在原子操作中将文件切换到覆盖目标之前准备文件（workdir必须与upperdir在同一文件系统上）。