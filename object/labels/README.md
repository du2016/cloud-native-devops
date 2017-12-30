Labels是连结到例如像pod这样的对象的键值对. 标签是意在被用来指定那些对象(有意义的并和用户有关的)的确认属性,但对核心系统不要直接使用隐含语义. 标签能够被用来去组织和去选择对象的子集.标签能够当对象创建的时候附加上去 并且随后能够在任何时候增加和修改. 每一个对象能够有一组明确的键值对标签.每一个键必须对给定的对象是独一无二的。

标签能够使用户去映射他们自己的组织结构在系统对象上以松耦合的方式并且无需客户去存储这些映射

通过 kubectl get pods --namespace=test test -o template --template="{{ .metadata.labels}}"

通过api 

相等依赖
?labelSelector=environment%3Dproduction,tier%3Dfrontend
kubectl get pods -l environment=production,tier=frontend

匹配依赖
?labelSelector=environment+in+%28production%2Cqa%29%2Ctier+in+%28frontend%29
kubectl get pods -l 'environment in (production),tier in (frontend)'