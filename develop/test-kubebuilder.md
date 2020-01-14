# kubebuilder

```
kubebuilder init --domain rocdu.top
kubebuilder create api --group networks --version v1 --kind Ip
kubebuilder create webhook --group networks --version v1 --kind Ip --defaulting --programmatic-validation
```

# 需要编辑文件

```
api/v1/ip_types.go # 结构定义
api/v1/ip_webhook.go # 校验逻辑
controllers/ip_controller.go # 监听逻辑
```

##  Validate逻辑
```
func (r *Ip) ValidateCreate() error {
	iplog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	if r.Spec.Foo=="" {
		return errors.New("can not nil")
	}
	return nil
}
```

## Reconcile 逻辑
```
func (r *IpReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("ip", req.NamespacedName)

	// your logic here
	var ip  v1.Ip
	if err:=r.Get(ctx,req.NamespacedName,&ip);err!=nil {
		log.Println(err)
	}else {
		log.Println(ip.Spec.Foo)
	}


	return ctrl.Result{}, nil
}
```