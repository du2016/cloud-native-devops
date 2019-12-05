prestop和prestart是pod声明周期的钩子

# 介绍

- PostStart 创建容器后，该挂钩立即执行。但是，不能保证挂钩将在容器ENTRYPOINT之前执行。没有参数传递给处理程序。
- PreStop 在容器终止之前立即执行，

PostStart因为不能保证执行顺序，在实际使用中很少用到
# 使用方式

- 脚本 内存小号计入pod
- http请求

## 脚本

我们定义了一个preStop用于执行aa.sh脚本

```
    lifecycle:
      preStop:
        exec:
          command:
          - /bin/bash
          - -c
          - aa.sh
```

## httpget

```
    lifecycle:
      preStop:
        httpGet:
          path: /admin/shutdown
          port: 8081
```