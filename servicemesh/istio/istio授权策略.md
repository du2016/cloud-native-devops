```json
{
  "name": "envoy.filters.http.rbac",
  "typed_config": {
    "@type": "type.googleapis.com/envoy.config.filter.http.rbac.v2.RBAC",
    "rules": {
      "policies": {
        "ns[foo]-policy[httpbin-deny-allowget]-rule[0]": {
          "permissions": [
            {
              "and_rules": {
                "rules": [
                  {
                    "or_rules": {
                      "rules": [
                        {
                          "header": {
                            "name": ":method",
                            "exact_match": "GET"
                          }
                        }
                      ]
                    }
                  }
                ]
              }
            }
          ],
          "principals": [
            {
              "and_ids": {
                "ids": [
                  {
                    "or_ids": {
                      "ids": [
                        {
                          "metadata": {
                            "filter": "istio_authn",
                            "path": [
                              {
                                "key": "source.principal"
                              }
                            ],
                            "value": {
                              "string_match": {
                                "exact": "cluster.local/ns/foo/sa/sleep"
                              }
                            }
                          }
                        }
                      ]
                    }
                  }
                ]
              }
            }
          ]
        }
      }
    }
  }
}
```

- permissions 必须，定义角色的权限集。 每个权限都与OR语义匹配。 要匹配此策略的所有操作，应使用any字段设置为true的单个Permission。

- principals 必须， 根据`action` 分配/拒绝 角色的principals集。 每个principals都与OR语义匹配。 为了匹配此策略的所有下游，应使用any字段设置为true的单个Principal。

- condition 指定访问控制条件的可选符号表达式。 该条件与permissions和principals组合为具有AND语义的子句。


func buildFilter(in *plugin.InputParams, mutable *networking.MutableObjects)