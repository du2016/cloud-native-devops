> 管理在单个服务网格中的多个Kubernetes集群上运行的微服务

![](https://miro.medium.com/max/1400/1*Xpbg9kPSIOv-flCUChv5xw.jpeg)

想象一下，您在一个典型的企业中工作，有多个团队一起工作，提供组成一个应用程序的单独软件。您的团队遵循微服务架构，并拥有由多个Kubernetes集群组成的广泛基础架构。

随着微服务分布在多个集群中，您需要设计一个解决方案来集中管理所有微服务。幸运的是，您正在使用Istio，提供此解决方案只是另一个配置更改。

像Istio这样的服务网格技术可帮助您安全地发现和连接分布在多个集群和环境中的微服务。

# 架构

Istio通过以下组件提供跨集群服务发现：

- istio核心DNS：每个Istio控制平面都附带一个核心DNS。Istio使用它来发现在全局范围内定义的服务。
例如，如果群集1上托管的微服务需要连接到群集2上托管的另一个微服务，
则需要在Istio Core DNS上为集群2上运行的微服务创建全局条目。

- 根CA：由于Istio要求在单独的群集上运行的服务之间建立mTLS连接，因此您需要使用共享的根CA为两个群集生成中间CA证书。
由于中间证书共享相同的根CA，因此在不同群集上运行的微服务之间建立了信任。

- Istio入口网关：群集间通信通过入口网关进行，并且服务之间没有直接连接。因此，您需要确保Ingress Gateway是可发现的，并且所有群集都可以连接到它。

![集群间通信](https://miro.medium.com/max/1362/1*AJMMpOukN4b5cbS8i1Kpww.png)

# 服务发现

Istio使用以下步骤促进服务发现：


- 所有集群上都有相同的控制平面，以提高高可用性。
- Kube DNS被存入Istio Core DNS，以提供全局服务发现。
- 用户通过Istio Core DNS中的ServiceEntries以name.namespace.global格式定义到远程服务的路由。
- 源Sidecar使用全局Core DNS条目将流量路由到目标Istio Ingress Gateway。
- 目标Istio Ingress网关会将流量路由到正确的微服务Pod。

# 先决条件

本文假定您对Kubernetes和Istio的工作原理有基本的了解,要进行动手演示，请确保：

- 您至少有两个运行版本1.14、1.15或1.16的Kubernetes集群。
- 您有权在群集中安装和配置Istio。
- 您在两个Kubernetes集群上都具有集群管理员访问权限。
- 入口网关可通过网络负载平衡器或类似配置访问其他群集。不需要平面网络。

