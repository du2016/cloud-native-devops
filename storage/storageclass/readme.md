#

# DefaultStorageClass准入控制器

这个插件观察不指定 storage class 字段的 PersistentVolumeClaim 对象的创建，
并自动向它们添加默认的 storage class 。
这样，不指定 storage class 字段的用户根本无需关心它们，它们将得到默认的 storage class 。

当没有配置默认 storage class 时，这个插件不会执行任何操作。
当一个以上的 storage class 被标记为默认时，
它拒绝 PersistentVolumeClaim 创建并返回一个错误，
管理员必须重新检查 StorageClass 对象，并且只标记一个作为默认值。
这个插件忽略了任何 PersistentVolumeClaim 更新，它只对创建起作用